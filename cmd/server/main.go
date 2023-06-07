package main

import (
	"github.com/nktau/monitoring-service/internal/applayer"
	"github.com/nktau/monitoring-service/internal/httplayer"
	"github.com/nktau/monitoring-service/internal/storagelayer"
)

func main() {
	storeLayer := storagelayer.New()

	// create app layer
	appLayer := applayer.New(storeLayer)

	// create http layer
	httplayer.New(appLayer)

}
