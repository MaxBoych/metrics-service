package main

import (
	"flag"
	"os"
)

type Config struct {
	flagRunAddr string
}

func parseConfig() (config Config) {
	flag.StringVar(&config.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		config.flagRunAddr = envRunAddr
	}

	return
}
