package main

import (
	"github.com/MaxBoych/MetricsService/cmd/handlers"
	"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"
	"net/http"

	"github.com/MaxBoych/MetricsService/cmd/storage"
)

func main() {
	ms := &storage.MemStorage{}
	ms.Init()
	msHandler := &handlers.MetricsHandler{MS: ms}

	//router.Use(middleware.AllowContentType("text/plain"))

	router := chi.NewRouter()

	router.Get("/", msHandler.GetAllMetrics)
	router.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", msHandler.GetGaugeMetric)
		r.Get("/counter/{name}", msHandler.GetCounterMetric)
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", msHandler.UpdateGaugeMetric)
		r.Post("/counter/{name}/{value}", msHandler.UpdateCounterMetric)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handlers.BadRequest(w, r)
	})

	//router.NotFound(handlers.NotFound)

	//mux := http.NewServeMux()
	//mux.HandleFunc("/", handlers.UndefinedMetric)
	//mux.Handle("/update/gauge/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateGaugeMetric)))
	//mux.Handle("/update/counter/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateCounterMetric)))

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
