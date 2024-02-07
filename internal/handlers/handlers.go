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
	UpdateGauge(name string, new storage.Gauge) storage.Gauge
	UpdateCounter(name string, new storage.Counter) storage.Counter
	Count()
	GetGauge(name string) (string, bool)
	GetCounter(name string) (string, bool)
	GetAllMetrics() []string
}

type MetricsHandler struct {
	ms Repository
	db *storage.DBStorage
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func NewMetricsHandler(ms *storage.MemStorage, db *storage.DBStorage) (handler *MetricsHandler) {
	handler = &MetricsHandler{
		ms: ms,
		db: db,
	}
	return
}

func (o *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if accept := r.Header.Get("Accept"); accept == "html/text" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	metrics := o.ms.GetAllMetrics()
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

func (o *MetricsHandler) GetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	var value string
	var ok bool
	if value, ok = o.ms.GetGauge(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(value))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) GetCounterMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	var value string
	var ok bool
	if value, ok = o.ms.GetCounter(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(value))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) UpdateGaugeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	if value, err := strconv.ParseFloat(value, 64); err == nil {
		o.ms.UpdateGauge(name, storage.Gauge(value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (o *MetricsHandler) UpdateCounterMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	if value, err := strconv.ParseInt(value, 10, 64); err == nil {
		o.ms.UpdateCounter(name, storage.Counter(value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (o *MetricsHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
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
		g := float64(o.ms.UpdateGauge(metricName, storage.Gauge(*metrics.Value)))
		resp = models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &g,
		}
	} else if metricType == "counter" {
		c := int64(o.ms.UpdateCounter(metricName, storage.Counter(*metrics.Delta)))
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

func (o *MetricsHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
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
		gStr, ok := o.ms.GetGauge(metricName)
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
		cStr, ok := o.ms.GetCounter(metricName)
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

func (o *MetricsHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	err := o.db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
