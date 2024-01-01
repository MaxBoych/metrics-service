package handlers

import (
	//"fmt"
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"net/http"
	"strconv"
	"strings"
)

type MetricsHandler struct {
	MS storage.Repository
}

func UndefinedMetric(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(w, r)
	})
}

func (handler *MetricsHandler) ReceiveGaugeMetric(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	//w.Header().Set("Content-Type", )
	metricData := strings.Split(path, "/")
	//w.Write([]byte(strings.Join(metricData, "-")))
	//w.Write([]byte(strconv.Itoa(len(metricData))))
	if len(metricData) != 5 || !handler.MS.ContainsGauge(metricData[3]) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if value, err := strconv.ParseFloat(metricData[4], 64); err == nil {
		handler.MS.UpdateGauge(metricData[3], storage.Gauge(value))
		w.WriteHeader(http.StatusOK)
		//w.Write([]byte(fmt.Sprintf("Обновлена метрика %s = %f\n", metricData[3], value)))
		//fmt.Println(fmt.Sprintf("Обновлена метрика %s = %f\n", metricData[3], value))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		//w.Write([]byte("hereBad\n"))
	}
}

func (handler *MetricsHandler) ReceiveCounterMetric(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	metricData := strings.Split(path, "/")
	//w.Write([]byte(strings.Join(metricData, "-")))
	//w.Write([]byte(strconv.Itoa(len(metricData))))
	if len(metricData) != 5 || !handler.MS.ContainsCounter(metricData[3]) {
		w.WriteHeader(http.StatusNotFound)
		//fmt.Println("StatusNotFound")
		return
	}

	if value, err := strconv.ParseInt(metricData[4], 10, 64); err == nil {
		handler.MS.UpdateCounter(metricData[3], storage.Counter(value))
		w.WriteHeader(http.StatusOK)
		//fmt.Println("StatusOK")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		//fmt.Println("StatusBadRequest")
	}
}
