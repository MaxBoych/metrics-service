package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	config := parseConfig()

	var ms *storage.MemStorage
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		updateMetrics(ms, config)
	}()
	go func() {
		defer wg.Done()
		sendMetrics(ms, config)
	}()

	wg.Wait()
}

func updateMetrics(ms *storage.MemStorage, config Config) {
	var stats runtime.MemStats
	ms = storage.NewMemStorage()

	for {
		time.Sleep(time.Duration(config.flagPollInterval) * time.Second)

		runtime.ReadMemStats(&stats)

		ms.Mu.Lock()
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
		ms.Mu.Unlock()
	}
}

func sendMetrics(ms *storage.MemStorage, config Config) {
	for {
		time.Sleep(time.Duration(config.flagReportInterval) * time.Second)

		var gaugesCopy map[string]storage.Gauge

		ms.Mu.Lock()
		gaugesCopy = make(map[string]storage.Gauge, len(ms.Gauges))
		for k, v := range ms.Gauges {
			gaugesCopy[k] = v
		}
		ms.Mu.Unlock()

		for key, value := range gaugesCopy {
			url := fmt.Sprintf("http://%s/update/gauge/%s/%s", config.flagRunAddr, key, fmt.Sprint(value))
			response, err := http.Post(url, "text/plain", nil)
			if err != nil {
				log.Printf("Error sending POST request: %v\n", err)
				continue
			}
			err = response.Body.Close()
			if err != nil {
				log.Printf("Error closing response body: %v\n", err)
			}
		}
	}
}
