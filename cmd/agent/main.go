package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	ms := memory.NewMemStorage()
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

func updateMetrics(ms *memory.MemStorage, config Config) {
	var stats runtime.MemStats

	for {
		time.Sleep(time.Duration(config.pollInterval) * time.Second)

		runtime.ReadMemStats(&stats)

		ms.Mu.Lock()
		ms.Gauges["Alloc"] = models.Gauge(stats.Alloc)
		ms.Gauges["BuckHashSys"] = models.Gauge(stats.BuckHashSys)
		ms.Gauges["Frees"] = models.Gauge(stats.Frees)
		ms.Gauges["GCCPUFraction"] = models.Gauge(stats.GCCPUFraction)
		ms.Gauges["GCSys"] = models.Gauge(stats.GCSys)
		ms.Gauges["HeapAlloc"] = models.Gauge(stats.HeapAlloc)
		ms.Gauges["HeapIdle"] = models.Gauge(stats.HeapIdle)
		ms.Gauges["HeapInuse"] = models.Gauge(stats.HeapInuse)
		ms.Gauges["HeapObjects"] = models.Gauge(stats.HeapObjects)
		ms.Gauges["HeapReleased"] = models.Gauge(stats.HeapReleased)
		ms.Gauges["HeapSys"] = models.Gauge(stats.HeapSys)
		ms.Gauges["LastGC"] = models.Gauge(stats.LastGC)
		ms.Gauges["NextGC"] = models.Gauge(stats.NextGC)
		ms.Gauges["NumForcedGC"] = models.Gauge(stats.NumForcedGC)
		ms.Gauges["NumGC"] = models.Gauge(stats.NumGC)
		ms.Gauges["PauseTotalNs"] = models.Gauge(stats.PauseTotalNs)
		ms.Gauges["Mallocs"] = models.Gauge(stats.Mallocs)
		ms.Gauges["TotalAlloc"] = models.Gauge(stats.TotalAlloc)
		ms.Gauges["OtherSys"] = models.Gauge(stats.OtherSys)
		ms.Gauges["Sys"] = models.Gauge(stats.Sys)
		ms.Gauges["MCacheInuse"] = models.Gauge(stats.MCacheInuse)
		ms.Gauges["MCacheSys"] = models.Gauge(stats.MCacheSys)
		ms.Gauges["MSpanInuse"] = models.Gauge(stats.MSpanInuse)
		ms.Gauges["MSpanSys"] = models.Gauge(stats.MSpanSys)
		ms.Gauges["StackInuse"] = models.Gauge(stats.StackInuse)
		ms.Gauges["StackSys"] = models.Gauge(stats.StackSys)
		ms.Gauges["Lookups"] = models.Gauge(stats.Lookups)
		ms.Gauges["RandomValue"] = models.Gauge(rand.Float64())
		ms.Mu.Unlock()
	}
}

func sendMetrics(ms *memory.MemStorage, config Config) {
	for {
		time.Sleep(time.Duration(config.reportInterval) * time.Second)

		var gaugesCopy map[string]models.Gauge

		ms.Mu.RLock()
		gaugesCopy = make(map[string]models.Gauge, len(ms.Gauges))
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

func sendMetricsJSON(ms *memory.MemStorage, config Config) {
	for {
		time.Sleep(time.Duration(config.reportInterval) * time.Second)

		var gaugesCopy map[string]models.Gauge

		ms.Mu.RLock()
		gaugesCopy = make(map[string]models.Gauge, len(ms.Gauges))
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
