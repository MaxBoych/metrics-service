package main

import (
	"context"
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/gzip"
	"github.com/MaxBoych/MetricsService/internal/handlers"
	"github.com/MaxBoych/MetricsService/internal/logger"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v4"
)

func main() {
	ms, msHandler := initState()
	defer ms.CloseDB()
	runServer(ms, msHandler)
}

func initState() (ms *storage.MemStorage, msHandler *handlers.MetricsHandler) {
	ms = storage.NewMemStorage()
	msHandler = handlers.NewMetricsHandler(ms)

	if err := logger.Initialize("INFO"); err != nil {
		fmt.Printf("logger init error: %v\n", err)
	}

	return
}

func runServer(ms *storage.MemStorage, msHandler *handlers.MetricsHandler) {
	router := chi.NewRouter()
	router.Use(logger.MiddlewareLogger)
	//router.Use(gzip.MiddlewareGzip)
	router.Use(middleware.Compress(5, "gzip"))
	router.Use(gzip.MiddlewareGzipReader)

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

	cfg := configure(ms)

	logger.Log.Info("running server", zap.String("address", cfg.runAddr))
	if err := http.ListenAndServe(cfg.runAddr, router); err != nil {
		panic(err)
	}
}

func configure(ms *storage.MemStorage) (config Config) {
	config = parseConfig()
	ms.FilePath = config.fileStoragePath
	if config.restore {
		err := ms.LoadFromFile()
		if err != nil {
			logger.Log.Error("ERROR load from file", zap.String("error", err.Error()))
		}
	}
	if config.storeInterval > 0 {
		go func() {
			for {
				time.Sleep(time.Duration(config.storeInterval) * time.Second)

				err := ms.StoreToFile()
				if err != nil {
					logger.Log.Error("ERROR store to file", zap.String("error", err.Error()))
				}
			}
		}()
	} else if config.storeInterval == 0 {
		ms.AutoSave = true
	}

	logger.Log.Info("INFO config.databaseDSN != \"\"", zap.String("info", config.databaseDSN))
	if config.databaseDSN != "" {
		connectDB(config.databaseDSN, ms)
	}

	return
}

func connectDB(url string, ms *storage.MemStorage) {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		logger.Log.Error("Unable to connect to database", zap.String("err", err.Error()))
		return
	}
	logger.Log.Info("connecting to database", zap.String("address", url))
	ms.SetDB(conn)
}
