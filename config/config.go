package main

import (
	"errors"
	"flag"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/file"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/postgres"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

type Config struct {
	runAddr         string
	fileStoragePath string
	restore         bool
	storeInterval   int
	databaseDSN     string
}

func NewConfig() *Config {
	return &Config{}
}

func (o *Config) parseConfig() {
	flag.StringVar(&o.runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.fileStoragePath, "f", "/tmp/metrics-db.json", "file path to store metrics on the disk")
	flag.BoolVar(&o.restore, "r", true, "whether to load metrics from a file when the server starts")
	flag.IntVar(&o.storeInterval, "i", 300, "time interval to store metrics on the disk")
	flag.StringVar(&o.databaseDSN, "d", "", "database address to connect")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		o.runAddr = envRunAddr
	}
	if envFileStoragePath := os.Getenv("ADDRESS"); envFileStoragePath != "" {
		o.fileStoragePath = envFileStoragePath
	}
	if envRestore, err := strconv.ParseBool(os.Getenv("ADDRESS")); err == nil {
		o.restore = envRestore
	}
	if envStoreInterval, err := strconv.Atoi(os.Getenv("STORE_INTERVAL")); err == nil {
		o.storeInterval = envStoreInterval
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		o.databaseDSN = envDatabaseDSN
	}

	return
}

func (o *Config) configureFS(fs *file.FileStorage) error {
	if o.fileStoragePath != "" {
		fs.SetConfigValues(o.fileStoragePath, o.storeInterval == 0)

		if o.restore {
			err := fs.LoadFromFile()
			if err != nil {
				logger.Log.Error("ERROR load from file", zap.String("error", err.Error()))
				return err
			}
		}
		if o.storeInterval > 0 {
			go func() {
				for {
					time.Sleep(time.Duration(o.storeInterval) * time.Second)

					err := fs.StoreToFile()
					if err != nil {
						logger.Log.Error("ERROR store to file", zap.String("error", err.Error()))
					}
				}
			}()
		}
	}

	return nil
}

func (o *Config) configureDB(db *postgres.PGStorage) error {
	logger.Log.Info("INFO config.databaseDSN != \"\"", zap.String("info", o.databaseDSN))
	if o.databaseDSN != "" {
		err := db.Connect(o.databaseDSN)
		if err != nil {
			logger.Log.Info("ERROR cannot connect to DB", zap.String("err", err.Error()))
			return err
		}
		return nil
	} else {
		return errors.New("database DSN is empty")
	}
}
