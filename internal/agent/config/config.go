package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type config struct {
	ServerURL      string
	ReportInterval int
	PollInterval   int
}

func New() config {
	cfg := config{}
	cfg.parseFlags()
	cfg.parseEnv()
	cfg.ServerURL = fmt.Sprintf("http://%s", cfg.ServerURL)
	return cfg
}

func (cfg *config) parseFlags() {
	flag.StringVar(&cfg.ServerURL, "a", "localhost:8080", "endpoint of monitoring-service server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "frequency of polling metrics from the runtime package in seconds")
	flag.Parse()
}

func (cfg *config) parseEnv() {
	if envServerURL := os.Getenv("ADDRESS"); envServerURL != "" {
		cfg.ServerURL = envServerURL
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		reportInterval, err := strconv.Atoi(envReportInterval)
		if err == nil {
			cfg.ReportInterval = reportInterval
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		pollInterval, err := strconv.Atoi(envPollInterval)
		if err == nil {
			cfg.PollInterval = pollInterval
		}
	}
}
