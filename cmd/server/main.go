package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/handlers"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	ms := storage.NewMemStorage()
	msHandler := handlers.NewMetricsHandler(ms)

	router := chi.NewRouter()

	router.Get("/", msHandler.GetAllMetrics)
	router.Route("/value", func(r chi.Router) {

		r.Get("/gauge/{name}", msHandler.GetGaugeMetric)
		r.Get("/counter/{name}", msHandler.GetCounterMetric)

		r.NotFound(handlers.BadRequest)
	})

	router.Route("/update", func(r chi.Router) {

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
	fmt.Printf("server running on %s\n", config.flagRunAddr)
	err := http.ListenAndServe(config.flagRunAddr, router)
	if err != nil {
		panic(err)
	}
	fmt.Printf("server running on %s\n", config.flagRunAddr)
}
