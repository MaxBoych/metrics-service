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

	router := chi.NewRouter()
	//router.Use(middleware.AllowContentType("text/plain"))

	router.Get("/", msHandler.GetAllMetrics)
	router.Get("/value/gauge/{name}", msHandler.GetGaugeMetric)
	router.Get("/value/counter/{name}", msHandler.GetCounterMetric)

	router.Post("/update/gauge/{name}/{value}", msHandler.UpdateGaugeMetric)
	router.Post("/update/counter/{name}/{value}", msHandler.UpdateCounterMetric)

	router.NotFound(handlers.NotFound)

	//mux := http.NewServeMux()
	//mux.HandleFunc("/", handlers.UndefinedMetric)
	//mux.Handle("/update/gauge/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateGaugeMetric)))
	//mux.Handle("/update/counter/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateCounterMetric)))

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
