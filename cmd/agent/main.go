package main

import (
	"github.com/nktau/monitoring-service/internal/agent"
)

func main() {
	parseFlags()
	agent := agent.New()
	agent.Start("http://"+flagServerURL, flagReportInterval, flagPollInterval)

}
