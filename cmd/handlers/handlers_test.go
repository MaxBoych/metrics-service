package handlers

import (
	"github.com/MaxBoych/MetricsService/cmd/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const url = "http://localhost:8080/update/gauge/"

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
			endpoint: "HeapIdle/12345",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #1: Not Found",
			endpoint: "12345",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #2: Not Found",
			endpoint: "HeapIdleee/12345",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:     "fail test #3: Bad Request",
			endpoint: "HeapIdle/12qwerty345",
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

			handler := Middleware(http.HandlerFunc(msHandler.ReceiveMetric))
			handler.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, test.want.code, result.StatusCode)
			assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestMetricsHandler_ReceiveMetric(t *testing.T) {

}
