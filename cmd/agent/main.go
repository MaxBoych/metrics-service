package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/models"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	ms := storage.NewMemStorage()
	config := parseConfig()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		updateMetrics(ms, config)
	}()
	go func() {
		defer wg.Done()
		sendMetrics(ms, config)
		sendMetricsJSON(ms, config)
	}()

	wg.Wait()
}

func updateMetrics(ms *storage.MemStorage, config Config) {
	var stats runtime.MemStats

	for {
		time.Sleep(time.Duration(config.pollInterval) * time.Second)

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
		time.Sleep(time.Duration(config.reportInterval) * time.Second)

		var gaugesCopy map[string]storage.Gauge

		ms.Mu.RLock()
		gaugesCopy = make(map[string]storage.Gauge, len(ms.Gauges))
		for k, v := range ms.Gauges {
			gaugesCopy[k] = v
		}
		ms.Mu.RUnlock()

		for key, value := range gaugesCopy {
			url := fmt.Sprintf("http://%s/update/gauge/%s/%s", config.runAddr, key, fmt.Sprint(value))
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

func sendMetricsJSON(ms *storage.MemStorage, config Config) {
	for {
		time.Sleep(time.Duration(config.reportInterval) * time.Second)

		var gaugesCopy map[string]storage.Gauge

		ms.Mu.RLock()
		gaugesCopy = make(map[string]storage.Gauge, len(ms.Gauges))
		for k, v := range ms.Gauges {
			gaugesCopy[k] = v
		}
		ms.Mu.RUnlock()

		for key, value := range gaugesCopy {

			g := float64(value)
			metrics := models.Metrics{
				ID:    key,
				MType: "gauge",
				Value: &g,
			}
			jsonBody, err := json.Marshal(metrics)
			if err != nil {
				panic(err)
			}

			var b bytes.Buffer
			if config.useGzip {
				gz := gzip.NewWriter(&b)
				_, err = gz.Write(jsonBody)
				if err != nil {
					log.Printf("Error compressing JSON: %v\n", err)
					continue
				}
				gz.Close()
			} else {
				b.Write(jsonBody)
			}

			url := fmt.Sprintf("http://%s/update/", config.runAddr)
			request, err := http.NewRequest("POST", url, &b)
			if err != nil {
				log.Printf("Error creating request: %v\n", err)
				continue
			}

			request.Header.Set("Content-Type", "application/json")
			if config.useGzip {
				request.Header.Set("Content-Encoding", "gzip")
			}

			response, err := http.DefaultClient.Do(request)
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
