package applayer

import "github.com/nktau/monitoring-service/internal/storagelayer"

type app struct {
	store storagelayer.MemStorage
}

type App interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}
