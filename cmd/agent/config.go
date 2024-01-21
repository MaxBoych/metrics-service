package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	runAddr        string
	reportInterval int
	pollInterval   int
}

func parseConfig() (config Config) {
	flag.StringVar(&config.runAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&config.reportInterval, "r", 10, "frequency of sending metrics on the server")
	flag.IntVar(&config.pollInterval, "p", 2, "frequency of polling metrics from the 'runtime' package")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		config.runAddr = envRunAddr
	}
	if envReportInterval, err := strconv.Atoi(os.Getenv("REPORT_INTERVAL")); err == nil {
		config.reportInterval = envReportInterval
	}
	if envPollInterval, err := strconv.Atoi(os.Getenv("POLL_INTERVAL")); err == nil {
		config.pollInterval = envPollInterval
	}

	return
}
