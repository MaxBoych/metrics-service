package main

import (
	"github.com/MaxBoych/MetricsService/cmd/handlers"
	"net/http"

	"github.com/MaxBoych/MetricsService/cmd/storage"
)

func main() {
	ms := &storage.MemStorage{}
	ms.Init()
	msHandler := &handlers.MetricsHandler{MS: ms}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.UndefinedMetric)
	mux.Handle("/update/gauge/", handlers.Middleware(http.HandlerFunc(msHandler.ReceiveGaugeMetric)))
	mux.Handle("/update/counter/", handlers.Middleware(http.HandlerFunc(msHandler.ReceiveCounterMetric)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
