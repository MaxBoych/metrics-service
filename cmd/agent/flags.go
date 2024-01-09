package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics on the server")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of polling metrics from the 'runtime' package")
	flag.Parse()

	var err error
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		flagReportInterval, err = strconv.Atoi(envReportInterval)
		if err != nil {
			panic(err)
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		flagPollInterval, err = strconv.Atoi(envPollInterval)
		if err != nil {
			panic(err)
		}
	}
}
