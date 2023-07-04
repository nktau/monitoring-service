package main

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/httplayer"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	time "time"
)

func main() {
	config.Init()
	logger := utils.InitLogger()
	// create storage layer
	time := time.Now().Unix()
	fmt.Println(time)
	storeLayer := storagelayer.New(logger)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := httplayer.New(appLayer, logger)
	logger.Info("starting http server")
	if err := httpAPI.Start(config.Config.ListenAddress); err != nil {
		panic(err)
	}
}
