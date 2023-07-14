package main

import (
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/httplayer"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
)

type metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func main() {

	cfg := config.New()
	logger := utils.InitLogger()
	// create storage layer
	storeLayer := storagelayer.New(logger, cfg)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := httplayer.New(appLayer, logger)
	logger.Info("starting http server")
	if err := httpAPI.Start(cfg.ListenAddress); err != nil {
		panic(err)
	}
}
