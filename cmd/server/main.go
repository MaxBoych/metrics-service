package main

import (
	"github.com/MaxBoych/MetricsService/cmd/handlers"
	"github.com/go-chi/chi/v5"
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

	//router.NotFound(handlers.NotFound)

	//mux := http.NewServeMux()
	//mux.HandleFunc("/", handlers.UndefinedMetric)
	//mux.Handle("/update/gauge/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateGaugeMetric)))
	//mux.Handle("/update/counter/", handlers.Middleware(http.HandlerFunc(msHandler.UpdateCounterMetric)))

	parseFlags()
	err := http.ListenAndServe(flagRunAddr, router)
	if err != nil {
		panic(err)
	}
	//fmt.Println("Running server on", flagRunAddr)
}
