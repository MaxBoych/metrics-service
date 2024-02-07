package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/gzip"
	"github.com/MaxBoych/MetricsService/internal/handlers"
	"github.com/MaxBoych/MetricsService/internal/logger"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	ms, msHandler, db := initState()
	defer db.CloseDB()
	runServer(ms, msHandler, db)
}

func initState() (ms *storage.MemStorage, msHandler *handlers.MetricsHandler, db *storage.DBStorage) {
	ms = storage.NewMemStorage()
	db = storage.NewDBStorage()
	msHandler = handlers.NewMetricsHandler(ms, db)

	if err := logger.Initialize("INFO"); err != nil {
		fmt.Printf("logger init error: %v\n", err)
	}

	return
}

func runServer(ms *storage.MemStorage, msHandler *handlers.MetricsHandler, db *storage.DBStorage) {
	router := chi.NewRouter()
	router.Use(logger.MiddlewareLogger)
	router.Use(gzip.MiddlewareGzip)

	router.Get("/", msHandler.GetAllMetrics)
	router.Route("/value", func(r chi.Router) {

		r.Post("/", msHandler.GetMetricJSON)
		r.Get("/gauge/{name}", msHandler.GetGaugeMetric)
		r.Get("/counter/{name}", msHandler.GetCounterMetric)

		r.NotFound(handlers.BadRequest)
	})

	router.Route("/update", func(r chi.Router) {

		r.Post("/", msHandler.UpdateMetricJSON)
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/{name}/{value}", msHandler.UpdateGaugeMetric)
			r.NotFound(handlers.NotFound)
		})

		r.Route("/counter", func(r chi.Router) {
			r.Post("/{name}/{value}", msHandler.UpdateCounterMetric)
			r.NotFound(handlers.NotFound)
		})

		r.NotFound(handlers.BadRequest)
	})

	router.Get("/ping", msHandler.PingDB)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handlers.BadRequest(w, r)
	})

	fs := storage.NewFileStorage(ms)
	cfg := configure(fs, db)

	logger.Log.Info("running server", zap.String("address", cfg.runAddr))
	if err := http.ListenAndServe(cfg.runAddr, router); err != nil {
		panic(err)
	}
}

func configure(fs *storage.FileStorage, db *storage.DBStorage) (config Config) {
	config = parseConfig()
	fs.SetConfigValues(config.fileStoragePath, config.storeInterval == 0)

	if config.restore {
		err := fs.LoadFromFile()
		if err != nil {
			logger.Log.Error("ERROR load from file", zap.String("error", err.Error()))
		}
	}
	if config.storeInterval > 0 {
		go func() {
			for {
				time.Sleep(time.Duration(config.storeInterval) * time.Second)

				err := fs.StoreToFile()
				if err != nil {
					logger.Log.Error("ERROR store to file", zap.String("error", err.Error()))
				}
			}
		}()
	}

	logger.Log.Info("INFO config.databaseDSN != \"\"", zap.String("info", config.databaseDSN))
	if config.databaseDSN != "" {
		db.Connect(config.databaseDSN)
	}

	return
}
