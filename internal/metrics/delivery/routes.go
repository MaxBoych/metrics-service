package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetupRoutes() {
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
}
