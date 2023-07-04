package storagelayer

import (
	"encoding/json"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/config"
	"go.uber.org/zap"
	"os"
	"time"
)

type memStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	logger  *zap.Logger
}

type MemStorage interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() (map[string]float64, map[string]int64)
}

func New(logger *zap.Logger) *memStorage {
	mem := &memStorage{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
		logger:  logger,
	}
	if config.Config.StoreInterval != 0 && config.Config.FileStoragePath != "" {
		go mem.writeToDiskWithStoreInterval()
	}
	if !config.Config.Restore {
		if _, err := os.Stat(config.Config.FileStoragePath); err == nil {
			os.Remove(config.Config.FileStoragePath)
		}
	}
	return mem

}

func (mem *memStorage) writeToDiskWithStoreInterval() error {
	count := 0
	fmt.Println("writeToDisk")
	for {
		if count%config.Config.StoreInterval == 0 {
			err := mem.writeToDisk()
			if err != nil {
				return err
			}
		}
		count++
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (mem *memStorage) writeToDisk() error {
	file, err := os.OpenFile(config.Config.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		mem.logger.Error("writeToDisk can't open file to store data", zap.Error(err))
		file.Close()
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(mem)
	file.Close()
	return nil
}
