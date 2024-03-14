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
	useGzip        bool
	Key            string
	rateLimit      int
}

func parseConfig() (config Config) {
	flag.StringVar(&config.runAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&config.reportInterval, "r", 10, "frequency of sending metrics on the server")
	flag.IntVar(&config.pollInterval, "p", 2, "frequency of polling metrics from the 'runtime' package")
	flag.BoolVar(&config.useGzip, "g", false, "whether to use gzip compression")
	flag.StringVar(&config.Key, "k", "", "hash key")
	flag.IntVar(&config.rateLimit, "l", 1, "limit of simultaneous outgoing requests")
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
	if envUseGzip, err := strconv.ParseBool(os.Getenv("POLL_INTERVAL")); err == nil {
		config.useGzip = envUseGzip
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		config.Key = envKey
	}
	if envRateLimit, err := strconv.Atoi(os.Getenv("RATE_LIMIT")); err == nil {
		config.rateLimit = envRateLimit
	}

	return
}
