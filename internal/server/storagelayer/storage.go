package storagelayer

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/config"
	"go.uber.org/zap"
	"os"
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
	if config.Config.Restore && config.Config.FileStoragePath != "" {
		mem.readFromDisk()
		fmt.Println(mem)
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
