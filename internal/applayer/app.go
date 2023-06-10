package applayer

import (
	"github.com/nktau/monitoring-service/internal/storagelayer"
)

var MetricNotFound = storagelayer.MetricNotFound

type app struct {
	store storagelayer.MemStorage
}

type App interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() (map[string]float64, map[string]int64)
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}
