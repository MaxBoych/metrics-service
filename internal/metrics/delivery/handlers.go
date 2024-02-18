package delivery

import (
	"encoding/json"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type MetricsHandler struct {
	useCase metrics.UseCase
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

func (o *MetricsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if accept := r.Header.Get("Accept"); accept == "html/text" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	resp, err := o.useCase.GetAll(ctx)
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
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	params := models.Metrics{ID: name}
	gauge, err := o.useCase.GetGauge(ctx, params)
	if gauge == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = w.Write([]byte(gauge.String()))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) GetCounterMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	params := models.Metrics{ID: name}
	counter, err := o.useCase.GetCounter(ctx, params)
	if counter == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = w.Write([]byte(counter.String()))
	if err != nil {
		panic(err)
	}
}

func (o *MetricsHandler) UpdateGaugeMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/plain")

	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		params := models.Metrics{ID: name, Value: &value}
		_, err = o.useCase.UpdateGauge(ctx, params)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (o *MetricsHandler) UpdateCounterMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Info("UpdateMetricJSON", zap.String("metric", metric.String()))

	metricName := metric.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metric.MType
	var resp models.Metrics
	if metricType == models.GaugeMetricName {
		params := models.Metrics{ID: metricName, Value: metric.Value}
		_, err = o.useCase.UpdateGauge(ctx, params)
		resp = models.Metrics{
			ID:    metricName,
			MType: models.GaugeMetricName,
			Value: metric.Value,
		}
	} else if metricType == models.CounterMetricName {
		params := models.Metrics{ID: metricName, Delta: metric.Delta}
		_, err = o.useCase.UpdateCounter(ctx, params)
		resp = models.Metrics{
			ID:    metricName,
			MType: models.CounterMetricName,
			Delta: metric.Delta,
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
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&metric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Info("GetMetricJSON", zap.String("metric", metric.String()))

	metricName := metric.ID
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := metric.MType
	var resp models.Metrics
	if metricType == models.GaugeMetricName {
		params := models.Metrics{ID: metricName}
		gauge, _ := o.useCase.GetGauge(ctx, params)
		if gauge == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		v := float64(*gauge)
		resp = models.Metrics{
			ID:    metricName,
			MType: models.GaugeMetricName,
			Value: &v,
		}
	} else if metricType == models.CounterMetricName {
		params := models.Metrics{ID: metricName}
		counter, _ := o.useCase.GetCounter(ctx, params)
		if counter == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		v := int64(*counter)
		resp = models.Metrics{
			ID:    metricName,
			MType: models.CounterMetricName,
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
	ctx := r.Context()

	err := o.useCase.Ping(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (o *MetricsHandler) UpdateMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var ms []models.Metrics
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&ms)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = o.useCase.UpdateMany(ctx, ms)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
