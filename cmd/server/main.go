package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/handlers"
	"github.com/MaxBoych/MetricsService/internal/logger"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	ms := storage.NewMemStorage()
	msHandler := handlers.NewMetricsHandler(ms)

	if err := logger.Initialize("INFO"); err != nil {
		fmt.Printf("logger init error: %v\n", err)
	}

	router := chi.NewRouter()
	router.Use(logger.MiddlewareLogger)

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

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handlers.BadRequest(w, r)
	})

	config := parseConfig()
	logger.Log.Info("running server", zap.String("address", config.flagRunAddr))
	if err := http.ListenAndServe(config.flagRunAddr, router); err != nil {
		panic(err)
	}
}
