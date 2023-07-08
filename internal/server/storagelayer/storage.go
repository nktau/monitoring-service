package storagelayer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
	"os"
	"time"
)

type memStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	logger  *zap.Logger
	config  config.Config
}

type MemStorage interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() (map[string]float64, map[string]int64)
	CheckDBConnection() error
}

var ErrMetricNotFound = errors.New("metric not found")

func New(logger *zap.Logger, config config.Config) *memStorage {
	mem := &memStorage{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
		logger:  logger,
		config:  config,
	}
	if config.Restore && config.FileStoragePath != "" {
		mem.readFromDisk()
		fmt.Println(mem)
	}
	if config.StoreInterval != 0 && config.FileStoragePath != "" {
		go mem.writeToDiskWithStoreInterval()
	}
	if !config.Restore {
		if _, err := os.Stat(config.FileStoragePath); err == nil {
			os.Remove(config.FileStoragePath)
		}
	}
	return mem
}

func (mem *memStorage) GetCounter(metricName string) (metricValue int64, err error) {
	value, ok := mem.Counter[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetGauge(metricName string) (metricValue float64, err error) {
	value, ok := mem.Gauge[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetAll() (map[string]float64, map[string]int64) {
	return mem.Gauge, mem.Counter
}

func (mem *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	mem.Gauge[metricName] = metricValue
	if mem.config.StoreInterval == 0 && mem.config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	mem.Counter[metricName] += metricValue
	if mem.config.StoreInterval == 0 && mem.config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) writeToDiskWithStoreInterval() error {
	count := 0
	for {
		if count%mem.config.StoreInterval == 0 {
			err := mem.writeToDisk()
			if err != nil {
				return err
			}
		}
		count++
		time.Sleep(1 * time.Second)
	}
}

func (mem *memStorage) writeToDisk() error {
	file, err := os.OpenFile(mem.config.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		mem.logger.Fatal("writeToDisk can't open file to store data", zap.Error(err))
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(mem)
	return nil
}

func (mem *memStorage) readFromDisk() error {
	err := json.Unmarshal([]byte(utils.GetLastLineWithSeek(mem.config.FileStoragePath)), mem)
	if err != nil {
		mem.logger.Error("readFromDisk unmarshal err", zap.Error(err))
		return err
	}
	return nil
}

func (mem *memStorage) CheckDBConnection() error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	return nil
}
