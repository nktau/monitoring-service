package main

import (
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/httplayer"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
)

func main() {

	cfg := config.New()
	logger := utils.InitLogger()
	// create storage layer
	storeLayer := storagelayer.New(logger, cfg)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := httplayer.New(appLayer, logger, cfg.HashKey)
	logger.Info("starting http server", zap.String("listenAddress", cfg.ListenAddress))
	if err := httpAPI.Start(cfg.ListenAddress); err != nil {
		panic(err)
	}
}
