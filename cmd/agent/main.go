package main

import (
	"fmt"
	. "github.com/MaxBoych/MetricsService/cmd/storage"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const pollInterval = 2
const reportInterval = 10

var ms MemStorage

func main() {
	go updateMetrics()
	go sendMetrics()

	select {}
}

func updateMetrics() {
	var stats runtime.MemStats
	ms = MemStorage{}
	ms.Gauges = map[string]Gauge{}
	for {
		time.Sleep(pollInterval * time.Second)

		runtime.ReadMemStats(&stats)

		ms.Gauges["Alloc"] = Gauge(stats.Alloc)
		ms.Gauges["BuckHashSys"] = Gauge(stats.BuckHashSys)
		ms.Gauges["Frees"] = Gauge(stats.Frees)
		ms.Gauges["GCCPUFraction"] = Gauge(stats.GCCPUFraction)
		ms.Gauges["GCSys"] = Gauge(stats.GCSys)
		ms.Gauges["HeapAlloc"] = Gauge(stats.HeapAlloc)
		ms.Gauges["HeapIdle"] = Gauge(stats.HeapIdle)
		ms.Gauges["HeapInuse"] = Gauge(stats.HeapInuse)
		ms.Gauges["HeapObjects"] = Gauge(stats.HeapObjects)
		ms.Gauges["HeapReleased"] = Gauge(stats.HeapReleased)
		ms.Gauges["HeapSys"] = Gauge(stats.HeapSys)
		ms.Gauges["LastGC"] = Gauge(stats.LastGC)
		ms.Gauges["NextGC"] = Gauge(stats.NextGC)
		ms.Gauges["NumForcedGC"] = Gauge(stats.NumForcedGC)
		ms.Gauges["NumGC"] = Gauge(stats.NumGC)
		ms.Gauges["PauseTotalNs"] = Gauge(stats.PauseTotalNs)
		ms.Gauges["Mallocs"] = Gauge(stats.Mallocs)
		ms.Gauges["TotalAlloc"] = Gauge(stats.TotalAlloc)
		ms.Gauges["OtherSys"] = Gauge(stats.OtherSys)
		ms.Gauges["Sys"] = Gauge(stats.Sys)
		ms.Gauges["MCacheInuse"] = Gauge(stats.MCacheInuse)
		ms.Gauges["MCacheSys"] = Gauge(stats.MCacheSys)
		ms.Gauges["MSpanInuse"] = Gauge(stats.MSpanInuse)
		ms.Gauges["MSpanSys"] = Gauge(stats.MSpanSys)
		ms.Gauges["StackInuse"] = Gauge(stats.StackInuse)
		ms.Gauges["StackSys"] = Gauge(stats.StackSys)
		ms.Gauges["Lookups"] = Gauge(stats.Lookups)
		ms.Gauges["RandomValue"] = Gauge(rand.Float64())
	}
}

func sendMetrics() {
	for {
		time.Sleep(reportInterval * time.Second)

		for key, value := range ms.Gauges {
			url := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%s", key, fmt.Sprint(value))
			response, err := http.Post(url, "text/plain", nil)
			if err != nil {
				panic(err)
			}
			err = response.Body.Close()
			if err != nil {
				panic(err)
			}
		}
	}
}
