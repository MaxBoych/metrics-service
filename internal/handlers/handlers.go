package handlers

import (
	"encoding/json"
	"github.com/MaxBoych/MetricsService/internal/models"
	"github.com/MaxBoych/MetricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Repository interface {
	Init()
	UpdateGauge(name string, new storage.Gauge) storage.Gauge
	UpdateCounter(name string, new storage.Counter) storage.Counter
	Count()
	GetGauge(name string) (string, bool)
	GetCounter(name string) (string, bool)
	GetAllMetrics() []string
	PingDB() error
}

type MetricsHandler struct {
	MS Repository
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func NewMetricsHandler(ms *storage.MemStorage) (handler *MetricsHandler) {
	handler = &MetricsHandler{
		MS: ms,
	}
	return
}

func (handler *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if accept := r.Header.Get("Accept"); accept == "html/text" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	metrics := handler.MS.GetAllMetrics()
	var result []byte
	for _, metric := range metrics {
		result = append(result, []byte(metric)...)
		result = append(result, '\n')
	}
	_, err := w.Write(result)
	if err != nil {
		panic(err)
	}
}

func (handler *MetricsHandler) GetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	var value string
	var ok bool
	if value, ok = handler.MS.GetGauge(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(value))
	if err != nil {
		panic(err)
	}
}

func (handler *MetricsHandler) GetCounterMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	var value string
	var ok bool
	if value, ok = handler.MS.GetCounter(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(value))
	if err != nil {
		panic(err)
	}
}

func (handler *MetricsHandler) UpdateGaugeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	if value, err := strconv.ParseFloat(value, 64); err == nil {
		handler.MS.UpdateGauge(name, storage.Gauge(value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *MetricsHandler) UpdateCounterMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	if value, err := strconv.ParseInt(value, 10, 64); err == nil {
		handler.MS.UpdateCounter(name, storage.Counter(value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *MetricsHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metrics models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metrics)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := metrics.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metrics.MType
	var resp models.Metrics
	if metricType == "gauge" {
		g := float64(handler.MS.UpdateGauge(metricName, storage.Gauge(*metrics.Value)))
		resp = models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &g,
		}
	} else if metricType == "counter" {
		c := int64(handler.MS.UpdateCounter(metricName, storage.Counter(*metrics.Delta)))
		resp = models.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &c,
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		panic(err)
	}
}

func (handler *MetricsHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metrics models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metrics)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := metrics.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metrics.MType
	var resp models.Metrics
	if metricType == "gauge" {
		gStr, ok := handler.MS.GetGauge(metricName)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		g, _ := strconv.ParseFloat(gStr, 64)
		resp = models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &g,
		}
	} else if metricType == "counter" {
		cStr, ok := handler.MS.GetCounter(metricName)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		c, _ := strconv.ParseInt(cStr, 10, 64)
		resp = models.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &c,
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		panic(err)
	}
}

func (handler *MetricsHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	err := handler.MS.PingDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
