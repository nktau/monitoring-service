package main

import (
	"github.com/nktau/monitoring-service/internal/agent/app"
	"github.com/nktau/monitoring-service/internal/agent/config"
)

func main() {
	cfg := config.New()
	agent := app.New()
	agent.Start(cfg.ServerURL, cfg.ReportInterval, cfg.PollInterval)
}
