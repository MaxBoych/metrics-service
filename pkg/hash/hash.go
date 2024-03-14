package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func Hash(value string, key string) string {
	hmacSHA256 := hmac.New(sha256.New, []byte(key))
	hmacSHA256.Write([]byte(value))
	hashBytes := hmacSHA256.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func MiddlewareHash(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		clientHash := r.Header.Get("HashSHA256")
		if clientHash != "" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			body := string(bodyBytes)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			hexHash := Hash(body, key)
			if clientHash != hexHash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		iw := newInterceptorWriter(w)

		next.ServeHTTP(iw, r)

		if clientHash != "" {
			hexHash := Hash(iw.body.String(), key)
			iw.Header().Set("HashSHA256", hexHash)
		}
	})
}

type interceptorWriter struct {
	w    http.ResponseWriter
	body *bytes.Buffer
}

func newInterceptorWriter(w http.ResponseWriter) *interceptorWriter {
	return &interceptorWriter{
		w:    w,
		body: bytes.NewBuffer(nil),
	}
}

func (w *interceptorWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.w.Write(b)
}

func (w *interceptorWriter) Header() http.Header {
	return w.w.Header()
}

func (w *interceptorWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}
