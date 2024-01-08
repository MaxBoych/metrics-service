package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

//const pollInterval = 2
//const reportInterval = 10

var ms storage.MemStorage

func main() {
	parseFlags()

	go updateMetrics()
	go sendMetrics()

	select {}
}

func updateMetrics() {
	var stats runtime.MemStats
	ms = storage.MemStorage{}
	ms.Gauges = map[string]storage.Gauge{}
	for {
		time.Sleep(time.Duration(flagPollInterval) * time.Second)

		runtime.ReadMemStats(&stats)

		ms.Gauges["Alloc"] = storage.Gauge(stats.Alloc)
		ms.Gauges["BuckHashSys"] = storage.Gauge(stats.BuckHashSys)
		ms.Gauges["Frees"] = storage.Gauge(stats.Frees)
		ms.Gauges["GCCPUFraction"] = storage.Gauge(stats.GCCPUFraction)
		ms.Gauges["GCSys"] = storage.Gauge(stats.GCSys)
		ms.Gauges["HeapAlloc"] = storage.Gauge(stats.HeapAlloc)
		ms.Gauges["HeapIdle"] = storage.Gauge(stats.HeapIdle)
		ms.Gauges["HeapInuse"] = storage.Gauge(stats.HeapInuse)
		ms.Gauges["HeapObjects"] = storage.Gauge(stats.HeapObjects)
		ms.Gauges["HeapReleased"] = storage.Gauge(stats.HeapReleased)
		ms.Gauges["HeapSys"] = storage.Gauge(stats.HeapSys)
		ms.Gauges["LastGC"] = storage.Gauge(stats.LastGC)
		ms.Gauges["NextGC"] = storage.Gauge(stats.NextGC)
		ms.Gauges["NumForcedGC"] = storage.Gauge(stats.NumForcedGC)
		ms.Gauges["NumGC"] = storage.Gauge(stats.NumGC)
		ms.Gauges["PauseTotalNs"] = storage.Gauge(stats.PauseTotalNs)
		ms.Gauges["Mallocs"] = storage.Gauge(stats.Mallocs)
		ms.Gauges["TotalAlloc"] = storage.Gauge(stats.TotalAlloc)
		ms.Gauges["OtherSys"] = storage.Gauge(stats.OtherSys)
		ms.Gauges["Sys"] = storage.Gauge(stats.Sys)
		ms.Gauges["MCacheInuse"] = storage.Gauge(stats.MCacheInuse)
		ms.Gauges["MCacheSys"] = storage.Gauge(stats.MCacheSys)
		ms.Gauges["MSpanInuse"] = storage.Gauge(stats.MSpanInuse)
		ms.Gauges["MSpanSys"] = storage.Gauge(stats.MSpanSys)
		ms.Gauges["StackInuse"] = storage.Gauge(stats.StackInuse)
		ms.Gauges["StackSys"] = storage.Gauge(stats.StackSys)
		ms.Gauges["Lookups"] = storage.Gauge(stats.Lookups)
		ms.Gauges["RandomValue"] = storage.Gauge(rand.Float64())
	}
}

func sendMetrics() {
	for {
		time.Sleep(time.Duration(flagReportInterval) * time.Second)

		for key, value := range ms.Gauges {
			url := fmt.Sprintf("http://%s/update/gauge/%s/%s", flagRunAddr, key, fmt.Sprint(value))
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
