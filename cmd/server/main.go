package main

import (
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/httplayer"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
)

func main() {
	// create storage layer
	storeLayer := storagelayer.New()
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := httplayer.New(appLayer)
	cfg := config.New()
	if err := httpAPI.Start(cfg.ListenAddress); err != nil {
		panic(err)
	}
}
