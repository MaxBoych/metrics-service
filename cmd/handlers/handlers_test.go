package handlers

import (
	//"fmt"
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//const urlUpdate = "http://localhost:8080/update/"
//const urlGet = "http://localhost:8080/value/"

func TestMetricsHandler_UpdateGaugeMetric(t *testing.T) {
	ms := &storage.MemStorage{}
	ms.Init()
	msHandler := &MetricsHandler{MS: ms}

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", msHandler.UpdateGaugeMetric)

	server := httptest.NewServer(router)
	defer server.Close()

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name     string
		endpoint string
		method   string
		want     want
	}{
		{
			name:     "UPDATE GAUGE pass test #1",
			method:   http.MethodPost,
			endpoint: "/update/gauge/HeapIdle/12345",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE GAUGE pass test #2",
			method:   http.MethodPost,
			endpoint: "/update/gauge/testGauge/100",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE GAUGE fail test #1: Not Found",
			method:   http.MethodPost,
			endpoint: "/update/gauge/12345",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE GAUGE fail test #2: Not Found",
			method:   http.MethodPost,
			endpoint: "/update/gauge/",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE GAUGE fail test #3: Bad Request",
			method:   http.MethodPost,
			endpoint: "/update/gauge/HeapIdle/12qwerty345",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE GAUGE fail test #4: Bad Request",
			method:   http.MethodPost,
			endpoint: "/update/gauge/testGauge/none",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = test.method
			request.URL = server.URL + test.endpoint
			//request.Header.Set("Content-Type", "text/plain")
			//t.Log(request.URL + "\n\n\n\n\n")

			response, err := request.Send()
			assert.NoErrorf(t, err, "Error making HTTP request")
			assert.Equal(t, test.want.code, response.StatusCode(), "Response code didn't match expected")
			assert.True(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/plain"))
		})
	}
}

func TestMetricsHandler_UpdateCounterMetric(t *testing.T) {
	ms := &storage.MemStorage{}
	ms.Init()
	msHandler := &MetricsHandler{MS: ms}

	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", msHandler.UpdateCounterMetric)

	server := httptest.NewServer(router)
	defer server.Close()

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name     string
		endpoint string
		method   string
		want     want
	}{
		{
			name:     "UPDATE COUNTER pass test #1",
			method:   http.MethodPost,
			endpoint: "/update/counter/testCounter/100",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE COUNTER fail test #1: Not Found",
			method:   http.MethodPost,
			endpoint: "/update/counter/",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE COUNTER fail test #2: Bad Request",
			method:   http.MethodPost,
			endpoint: "/update/counter/testCounter/12qwerty345",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:     "UPDATE COUNTER fail test #3: Bad Request",
			method:   http.MethodPost,
			endpoint: "/update/counter/testCounter/none",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = test.method
			request.URL = server.URL + test.endpoint
			//request.Header.Set("Content-Type", "text/plain")
			//t.Log(request.URL + "\n\n\n\n\n")

			response, err := request.Send()
			assert.NoErrorf(t, err, "Error making HTTP request")
			assert.Equal(t, test.want.code, response.StatusCode(), "Response code didn't match expected")
			assert.True(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/plain"))
		})
	}
}

func TestMetricsHandler_GetGaugeMetric(t *testing.T) {
	ms := &storage.MemStorage{}
	ms.Init()
	ms.Gauges["testGauge"] = 1155
	msHandler := &MetricsHandler{MS: ms}

	router := chi.NewRouter()
	router.Get("/value/gauge/{name}", msHandler.GetGaugeMetric)

	server := httptest.NewServer(router)
	defer server.Close()

	type want struct {
		code        int
		contentType string
		value       string
	}
	tests := []struct {
		name     string
		endpoint string
		method   string
		want     want
	}{
		{
			name:     "GET GAUGE pass test #1",
			method:   http.MethodGet,
			endpoint: "/value/gauge/testGauge",
			want: want{
				code:        200,
				contentType: "text/plain",
				value:       "1155",
			},
		},
		{
			name:     "GET GAUGE fail test #1: Not Found",
			method:   http.MethodGet,
			endpoint: "/value/gauge/testGauge123",
			want: want{
				code:        404,
				contentType: "text/plain",
				value:       "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = test.method
			request.URL = server.URL + test.endpoint
			//request.Header.Set("Content-Type", "text/plain")
			//t.Log(request.URL + "\n\n\n\n\n")

			response, err := request.Send()
			assert.NoErrorf(t, err, "Error making HTTP request")
			assert.Equal(t, test.want.code, response.StatusCode(), "Response code didn't match expected")
			assert.True(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/plain"))
			assert.Equal(t, test.want.value, string(response.Body()), "Response body didn't match expected")
		})
	}
}

func TestMetricsHandler_GetCounterMetric(t *testing.T) {
	ms := &storage.MemStorage{}
	ms.Init()
	ms.Counters["testCounter"] = 1177
	msHandler := &MetricsHandler{MS: ms}

	router := chi.NewRouter()
	router.Get("/value/counter/{name}", msHandler.GetCounterMetric)

	server := httptest.NewServer(router)
	defer server.Close()

	type want struct {
		code        int
		contentType string
		value       string
	}
	tests := []struct {
		name     string
		endpoint string
		method   string
		want     want
	}{
		{
			name:     "GET COUNTER pass test #1",
			method:   http.MethodGet,
			endpoint: "/value/counter/testCounter",
			want: want{
				code:        200,
				contentType: "text/plain",
				value:       "1177",
			},
		},
		{
			name:     "GET COUNTER fail test #1: Not Found",
			method:   http.MethodGet,
			endpoint: "/value/counter/testGauge123",
			want: want{
				code:        404,
				contentType: "text/plain",
				value:       "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = test.method
			request.URL = server.URL + test.endpoint
			//request.Header.Set("Content-Type", "text/plain")
			//t.Log(request.URL + "\n\n\n\n\n")

			response, err := request.Send()
			assert.NoErrorf(t, err, "Error making HTTP request")
			assert.Equal(t, test.want.code, response.StatusCode(), "Response code didn't match expected")
			assert.True(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/plain"))
			assert.Equal(t, test.want.value, string(response.Body()), "Response body didn't match expected")
		})
	}
}

/*func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	ms := &storage.MemStorage{}
	ms.Init()
	ms.Counters["testCounter"] = 1177
	ms.Counters["testGauge"] = 1155
	ms.Counters["HeapSys"] = 13
	msHandler := &MetricsHandler{MS: ms}

	router := chi.NewRouter()
	router.Get("/", msHandler.GetAllMetrics)

	server := httptest.NewServer(router)
	defer server.Close()

	type want struct {
		code        int
		contentType string
		value       string
	}
	tests := []struct {
		name     string
		endpoint string
		method   string
		want     want
	}{
		{
			name:     "GET ALL pass test #1",
			method:   http.MethodGet,
			endpoint: "",
			want: want{
				code:        200,
				contentType: "text/plain",
				value: "HeapReleased: 0\nRandomValue: 0\nFrees: 0\nGCCPUFraction: 0\nLastGC: 0\n" +
					"OtherSys: 0\nStackSys: 0\nAlloc: 0\nHeapAlloc: 0\nHeapInuse: 0\ntestGauge: 0\nGCSys: 0\n" +
					"NumForcedGC: 0\nTotalAlloc: 0\nStackInuse: 0\nHeapIdle: 0\nNextGC: 0\nNumGC: 0\nMallocs: 0\n" +
					"Sys: 0\nMSpanInuse: 0\nBuckHashSys: 0\nHeapSys: 0\nMCacheInuse: 0\nPauseTotalNs: 0\nMCacheSys: 0\n" +
					"Lookups: 0\nHeapObjects: 0\nMSpanSys: 0\nPollCount: 0\ntestCounter: 1177\ntestGauge: 1155\n" +
					"HeapSys: 13\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = test.method
			request.URL = server.URL + test.endpoint
			//request.Header.Set("Content-Type", "text/plain")
			//t.Log(request.URL + "\n\n\n\n\n")

			response, err := request.Send()
			assert.NoErrorf(t, err, "Error making HTTP request")
			assert.Equal(t, test.want.code, response.StatusCode(), "Response code didn't match expected")
			assert.True(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/plain"))
			assert.Equal(t, test.want.value, string(response.Body()), "Response body didn't match expected")
		})
	}
}*/
