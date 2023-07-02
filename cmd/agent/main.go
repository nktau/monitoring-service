package main

import (
	"github.com/nktau/monitoring-service/internal/agent/app"
	"github.com/nktau/monitoring-service/internal/agent/config"
	"github.com/nktau/monitoring-service/internal/server/utils"
)

func main() {
	cfg := config.New()
	logger := utils.InitLogger()
	agent := app.New(logger)
	logger.Info("starting agent")
	agent.Start(cfg.ServerURL, cfg.ReportInterval, cfg.PollInterval)
}
