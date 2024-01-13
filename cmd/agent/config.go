package main

import (
	"flag"
	//"github.com/caarlos0/env/v10"
	"os"
	"strconv"
)

type Config struct {
	runAddr        string `env:"ADDRESS"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	pollInterval   int    `env:"POLL_INTERVAL"`
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

	// библиотека "github.com/caarlos0/env/v10" с методом env.Parse() не работает (переменные окружения не считываются)
	//err := env.Parse(&config)
	//if err != nil {
	//	log.Fatalf("error parsing env vars: %v\n", err)
	//}

	return
}
