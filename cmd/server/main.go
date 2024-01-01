package main

import (
	. "github.com/MaxBoych/MetricsService/cmd/handlers"
	"net/http"

	. "github.com/MaxBoych/MetricsService/cmd/storage"
)

func main() {
	ms := &MemStorage{}
	ms.Init()
	msHandler := &MetricsHandler{MS: ms}

	mux := http.NewServeMux()
	mux.HandleFunc("/", UndefinedMetric)
	mux.Handle("/update/gauge/", Middleware(http.HandlerFunc(msHandler.ReceiveMetric)))
	//mux.Handle("/update/counter/", middleware(http.HandlerFunc(receiveCounterMetric)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
