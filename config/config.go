package config

import (
	"flag"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/file"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/postgres"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RunAddr         string
	FileStoragePath string
	Restore         bool
	StoreInterval   int
	DatabaseDSN     string
}

func NewConfig() *Config {
	return &Config{}
}

func (o *Config) ParseConfig() {
	flag.StringVar(&o.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.FileStoragePath, "f", "/tmp/metrics-db.json", "file path to store metrics on the disk")
	flag.BoolVar(&o.Restore, "r", true, "whether to load metrics from a file when the server starts")
	flag.IntVar(&o.StoreInterval, "i", 300, "time interval to store metrics on the disk")
	flag.StringVar(&o.DatabaseDSN, "d", "", "database address to connect")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		o.RunAddr = envRunAddr
	}
	if envFileStoragePath := os.Getenv("ADDRESS"); envFileStoragePath != "" {
		o.FileStoragePath = envFileStoragePath
	}
	if envRestore, err := strconv.ParseBool(os.Getenv("ADDRESS")); err == nil {
		o.Restore = envRestore
	}
	if envStoreInterval, err := strconv.Atoi(os.Getenv("STORE_INTERVAL")); err == nil {
		o.StoreInterval = envStoreInterval
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		o.DatabaseDSN = envDatabaseDSN
	}
}

func (o *Config) ConfigureDB() *postgres.PGStorage {
	logger.Log.Info("INFO config.DatabaseDSN", zap.String("DatabaseDSN", o.DatabaseDSN))
	db := postgres.NewDBStorage()
	if o.DatabaseDSN != "" {
		err := db.Connect(o.DatabaseDSN)
		if err != nil {
			logger.Log.Info("ERROR cannot connect to DB", zap.String("err", err.Error()))
			return nil
		}
		return db
	} else {
		logger.Log.Info("database DSN is empty")
		return nil
	}
}

func (o *Config) ConfigureFS(ms *memory.MemStorage) *file.FileStorage {
	fs := file.NewFileStorage(ms)
	if o.FileStoragePath != "" {
		fs.SetConfigValues(o.FileStoragePath, o.StoreInterval == 0)

		if o.Restore {
			err := fs.LoadFromFile()
			if err != nil {
				logger.Log.Error("ERROR load from file", zap.String("error", err.Error()))
				return nil
			}
		}

		if o.StoreInterval > 0 {
			go func() {
				for {
					time.Sleep(time.Duration(o.StoreInterval) * time.Second)

					err := fs.StoreToFile()
					if err != nil {
						logger.Log.Error("ERROR store to file", zap.String("error", err.Error()))
					}
				}
			}()
		}
	} else {
		logger.Log.Error("file path is empty")
		return nil
	}

	return fs
}

func (o *Config) ConfigureMS() *memory.MemStorage {
	return memory.NewMemStorage()
}
