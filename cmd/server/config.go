package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	runAddr         string
	fileStoragePath string
	restore         bool
	storeInterval   int
	databaseDSN     string
}

func parseConfig() (config Config) {
	flag.StringVar(&config.runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.fileStoragePath, "f", "/tmp/metrics-db.json", "file path to store metrics on the disk")
	flag.BoolVar(&config.restore, "r", true, "whether to load metrics from a file when the server starts")
	flag.IntVar(&config.storeInterval, "i", 300, "time interval to store metrics on the disk")
	flag.StringVar(&config.databaseDSN, "d", "", "database address to connect")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		config.runAddr = envRunAddr
	}
	if envFileStoragePath := os.Getenv("ADDRESS"); envFileStoragePath != "" {
		config.fileStoragePath = envFileStoragePath
	}
	if envRestore, err := strconv.ParseBool(os.Getenv("ADDRESS")); err == nil {
		config.restore = envRestore
	}
	if envStoreInterval, err := strconv.Atoi(os.Getenv("STORE_INTERVAL")); err == nil {
		config.storeInterval = envStoreInterval
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		config.databaseDSN = envDatabaseDSN
	}

	return
}
