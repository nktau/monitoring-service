package main

import (
	"github.com/nktau/monitoring-service/internal/applayer"
	"github.com/nktau/monitoring-service/internal/httplayer"
	"github.com/nktau/monitoring-service/internal/storagelayer"
)

func main() {
	// create storage layer
	storeLayer := storagelayer.New()
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := httplayer.New(appLayer)

	parseFlags()
	if err := httpAPI.Start(flagRunAddr); err != nil {
		panic(err)
	}
}
