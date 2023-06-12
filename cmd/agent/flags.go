package main

import (
	"flag"
)

var flagServerURL string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagServerURL, "a", "localhost:8080", "endpoint of monitoring-service server")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of polling metrics from the runtime package in seconds")
	flag.Parse()

}
