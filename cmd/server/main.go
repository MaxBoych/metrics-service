package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	gauge   gauge
	counter counter
}

type storage interface {
	replace(new gauge)
	add(new counter)
}

func (ms *MemStorage) replace(new gauge) {
	ms.gauge = new
}

func (ms *MemStorage) add(new counter) {
	ms.counter += new
}

var memStorage storage

func main() {
	memStorage = &MemStorage{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", undefinedMetric)
	mux.Handle("/update/gauge/", middleware(http.HandlerFunc(receiveGaugeMetric)))
	mux.Handle("/update/counter/", middleware(http.HandlerFunc(receiveCounterMetric)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func undefinedMetric(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(w, r)
	})
}

func receiveGaugeMetric(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	metricData := strings.Split(path, "/")
	//w.Write([]byte(strings.Join(metricData, "-")))
	//w.Write([]byte(strconv.Itoa(len(metricData))))
	if len(metricData) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if value, err := strconv.ParseFloat(metricData[4], 64); err == nil {
		memStorage.replace(gauge(value))
		w.WriteHeader(http.StatusOK)
		//w.Write([]byte("hereOK\n"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		//w.Write([]byte("hereBad\n"))
	}
}

func receiveCounterMetric(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	metricData := strings.Split(path, "/")
	//w.Write([]byte(strings.Join(metricData, "-")))
	//w.Write([]byte(strconv.Itoa(len(metricData))))
	if len(metricData) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if value, err := strconv.ParseInt(metricData[4], 10, 64); err == nil {
		memStorage.add(counter(value))
		w.WriteHeader(http.StatusOK)
		//w.Write([]byte("hereOK\n"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		//w.Write([]byte("hereBad\n"))
	}
}
