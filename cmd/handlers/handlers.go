package handlers

import (
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type MetricsHandler struct {
	MS storage.Repository
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (handler *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

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

//func Middleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != http.MethodPost {
//			w.WriteHeader(http.StatusMethodNotAllowed)
//			return
//		}
//
//		w.Header().Set("Content-Type", "text/plain")
//		next.ServeHTTP(w, r)
//	})
//}

//func (handler *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/plain")
//
//	metricType := chi.URLParam(r, "type")
//	if
//}

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

	//log.Printf("Request URL: %s", r.URL.String())
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	//log.Printf("UpdateGaugeMetric called with name: %s and value: %s", name, value)
	//if _, ok := handler.MS.GetGauge(name); !ok {
	//	w.WriteHeader(http.StatusNotFound)
	//	return
	//}

	//value := chi.URLParam(r, "value")
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
	//if _, ok := handler.MS.GetCounter(name); !ok {
	//	w.WriteHeader(http.StatusNotFound)
	//	return
	//}

	value := chi.URLParam(r, "value")
	if value, err := strconv.ParseInt(value, 10, 64); err == nil {
		handler.MS.UpdateCounter(name, storage.Counter(value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
