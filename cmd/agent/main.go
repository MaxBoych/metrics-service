package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/MaxBoych/MetricsService/pkg/hash"
	"github.com/MaxBoych/MetricsService/pkg/values"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	ms := memory.NewMemStorage()
	config := parseConfig()

	requests := make(chan *http.Request)
	errs := make(chan error)

	for w := 1; w <= config.rateLimit; w++ {
		go worker(requests, errs)
	}

	go func() {
		for err := range errs {
			log.Printf("Error doing request: %v\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		updateMetrics(ms, config)
	}()
	go func() {
		defer wg.Done()
		updateGopsutilMetrics(ms, config)
	}()
	go func() {
		defer wg.Done()
		sendMetrics(ms, config, requests)
		sendMetricsJSON(ms, config, requests)
		sendMany(ms, config, requests)
	}()

	wg.Wait()

	close(requests)
	close(errs)
}

func worker(requests <-chan *http.Request, errs chan<- error) {
	for r := range requests {
		if err := doRequest(r); err != nil {
			errs <- err
		}
	}
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

func updateGopsutilMetrics(ms *memory.MemStorage, config Config) {
	for {
		time.Sleep(time.Duration(config.pollInterval) * time.Second)

		vm, _ := mem.VirtualMemory()
		percent, _ := cpu.Percent(time.Second, false)

		ms.Gauges["TotalMemory"] = models.Gauge(vm.Total)
		ms.Gauges["FreeMemory"] = models.Gauge(vm.Free)
		ms.Gauges["CPUutilization1"] = models.Gauge(percent[0])
	}
}

func isConnectionRefused(err error) bool {
	var opError *net.OpError
	return errors.As(err, &opError)
}

func sendMetrics(ms *memory.MemStorage, config Config, requests chan<- *http.Request) {
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
			var err error
			var b bytes.Buffer
			//response, err = http.Post(url, "text/plain", nil)
			request, err := http.NewRequest("POST", url, &b)
			if err != nil {
				log.Printf("Error creating request: %v\n", err)
				continue
			}

			request.Header.Set("Content-Type", "text/plain")

			for _, interval := range values.RetryIntervals {
				requests <- request
				/*err = doRequest(request)
				if err == nil {
					break // успешный запрос
				} else if !isConnectionRefused(err) {
					break // проблема не на стороне сервера
				}

				log.Printf("Error sending POST request, retrying in %v: %v\n", interval, err)*/
				time.Sleep(interval)
			}

			/*if err != nil {
				log.Printf("Error after final retry: %v\n", err)
				continue
			}*/
		}
	}
}

func sendMetricsJSON(ms *memory.MemStorage, config Config, requests chan<- *http.Request) {
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
				MType: models.GaugeMetricName,
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
			if config.Key != "" {
				hexHash := hash.Hash(string(jsonBody), config.Key)
				request.Header.Set("HashSHA256", hexHash)
			}

			for _, interval := range values.RetryIntervals {
				requests <- request
				/*err = doRequest(request)
				if err == nil {
					break // успешный запрос
				} else if !isConnectionRefused(err) {
					break // проблема не на стороне сервера
				}

				log.Printf("Error sending POST request, retrying in %v: %v\n", interval, err)*/
				time.Sleep(interval)
			}

			/*if err != nil {
				log.Printf("Error after final retry: %v\n", err)
				continue
			}*/
		}
	}
}

func sendMany(ms *memory.MemStorage, config Config, requests chan<- *http.Request) {
	for {
		time.Sleep(time.Duration(config.reportInterval) * time.Second)

		var gaugesCopy map[string]models.Gauge

		ms.Mu.RLock()
		gaugesCopy = make(map[string]models.Gauge, len(ms.Gauges))
		for k, v := range ms.Gauges {
			gaugesCopy[k] = v
		}
		ms.Mu.RUnlock()

		var mcs []models.Metrics
		for key, value := range gaugesCopy {

			g := float64(value)
			m := models.Metrics{
				ID:    key,
				MType: models.GaugeMetricName,
				Value: &g,
			}
			mcs = append(mcs, m)
		}

		jsonBody, err := json.Marshal(mcs)
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
		if config.Key != "" {
			hexHash := hash.Hash(string(jsonBody), config.Key)
			request.Header.Set("HashSHA256", hexHash)
		}

		for _, interval := range values.RetryIntervals {
			requests <- request
			/*err = doRequest(request)
			if err == nil {
				break // успешный запрос
			} else if !isConnectionRefused(err) {
				break // проблема не на стороне сервера
			}

			log.Printf("Error sending POST request, retrying in %v: %v\n", interval, err)*/
			time.Sleep(interval)
		}

		/*if err != nil {
			log.Printf("Error after final retry: %v\n", err)
			continue
		}*/
	}
}

func doRequest(request *http.Request) error {
	response, err := http.DefaultClient.Do(request)
	if response != nil {
		defer response.Body.Close()
	}

	return err
}
