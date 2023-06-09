package main

import "github.com/nktau/monitoring-service/internal/agent"

func main() {
	store := agent.New()
	store.TriggerGetRuntimeMetric(2)
}
