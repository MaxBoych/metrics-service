package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	flagRunAddr        string `env:"ADDRESS"`
	flagReportInterval int    `env:"REPORT_INTERVAL"`
	flagPollInterval   int    `env:"POLL_INTERVAL"`
}

func parseConfig() (config Config) {
	flag.StringVar(&config.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&config.flagReportInterval, "r", 10, "frequency of sending metrics on the server")
	flag.IntVar(&config.flagPollInterval, "p", 2, "frequency of polling metrics from the 'runtime' package")
	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Fatalf("error parsing env vars: %v\n", err)
	}

	return
}
