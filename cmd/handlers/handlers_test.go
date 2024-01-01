package handlers

import (
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const url = "http://localhost:8080/update/"

func TestMiddleware(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name     string
		endpoint string
		want     want
	}{
		{
			name:     "pass test #1",
			endpoint: "gauge/HeapIdle/12345",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "pass test #2",
			endpoint: "counter/testCounter/100",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "pass test #3",
			endpoint: "gauge/testGauge/100",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #1: Not Found",
			endpoint: "gauge/12345",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #2: Not Found",
			endpoint: "gauge/HeapIdleee/12345",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #3: Not Found",
			endpoint: "counter/",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #4: Bad Request",
			endpoint: "gauge/HeapIdle/12qwerty345",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #5: Bad Request",
			endpoint: "gauge/testGauge/none",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, url+test.endpoint, nil)
			w := httptest.NewRecorder()

			ms := &storage.MemStorage{}
			ms.Init()
			msHandler := &MetricsHandler{MS: ms}

			var handler http.Handler
			if strings.HasPrefix(test.endpoint, "gauge") {
				handler = Middleware(http.HandlerFunc(msHandler.ReceiveGaugeMetric))
			} else {
				handler = Middleware(http.HandlerFunc(msHandler.ReceiveCounterMetric))
			}
			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.code, result.StatusCode)
			assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
