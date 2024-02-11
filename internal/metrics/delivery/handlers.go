package handlers

import (
	"context"
	"encoding/json"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type MetricsHandler struct {
	//repo metrics.Repository
	useCase metrics.UseCase
	//db      *postgres.PGStorage
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func NewMetricsHandler(useCase metrics.UseCase) (handler *MetricsHandler) {
	handler = &MetricsHandler{
		useCase: useCase,
	}
	return
}

func (o *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if accept := r.Header.Get("Accept"); accept == "html/text" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	/*ms := o.repo.GetAllMetrics(ctx)
	var result []byte
	for _, metric := range metrics {
		result = append(result, []byte(metric)...)
		result = append(result, '\n')
	}*/

	resp := o.useCase.GetAllMetrics(ctx)
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) GetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	params := models.Metrics{ID: name}
	var gauge *models.Gauge
	if gauge = o.useCase.GetGauge(ctx, params); gauge == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(gauge.String()))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) GetCounterMetric(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	params := models.Metrics{ID: name}
	var counter *models.Counter
	if counter = o.useCase.GetCounter(ctx, params); counter == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(counter.String()))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) UpdateGaugeMetric(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		params := models.Metrics{ID: name, Value: &value}
		_ = o.useCase.UpdateGauge(ctx, params)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (o *MetricsHandler) UpdateCounterMetric(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		params := models.Metrics{ID: name, Delta: &value}
		o.useCase.UpdateCounter(ctx, params)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (o *MetricsHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "application/json")

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := metric.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metric.MType
	var resp models.Metrics
	if metricType == "gauge" {
		params := models.Metrics{ID: metricName, Value: metric.Value}
		v := float64(*o.useCase.UpdateGauge(ctx, params))
		resp = models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &v,
		}
	} else if metricType == "counter" {
		params := models.Metrics{ID: metricName, Delta: metric.Delta}
		v := int64(*o.useCase.UpdateCounter(ctx, params))
		resp = models.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &v,
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
	ctx := context.Background()

	w.Header().Set("Content-Type", "application/json")

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := metric.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metric.MType
	var resp models.Metrics
	if metricType == "gauge" {
		params := models.Metrics{ID: metricName}
		gauge := o.useCase.GetGauge(ctx, params)
		if gauge == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		v := float64(*gauge)
		resp = models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &v,
		}
	} else if metricType == "counter" {
		params := models.Metrics{ID: metricName}
		counter := o.useCase.GetCounter(ctx, params)
		if counter == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		v := int64(*counter)
		resp = models.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &v,
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
	err := o.useCase.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
