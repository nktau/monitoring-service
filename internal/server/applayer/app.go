package applayer

import (
	storagelayer2 "github.com/nktau/monitoring-service/internal/server/storagelayer"
)

var ErrMetricNotFound = storagelayer2.ErrMetricNotFound

type app struct {
	store storagelayer2.MemStorage
}

type App interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() (map[string]float64, map[string]int64)
}

func New(store storagelayer2.MemStorage) *app {
	return &app{store: store}
}
