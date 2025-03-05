// Package middleware provides HTTP middleware components for the URL shortener service.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/learies/goShortener/internal/config/logger"
)

// ResponseWriter is a type alias for http.ResponseWriter.
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

// responseData is a struct that holds the status and size of the response.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter is a struct that wraps the http.ResponseWriter
// and adds logging functionality.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData  *responseData
	headerWritten bool
	mu            sync.Mutex
}

// Write overrides the default Write method to log the response size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	r.mu.Lock()
	if !r.headerWritten {
		r.responseData.status = http.StatusOK
		r.headerWritten = true
	}
	r.mu.Unlock()

	return size, err
}

// WriteHeader overrides the default WriteHeader method to log the status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.headerWritten {
		return
	}
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
	r.headerWritten = true
}

// WithLogging is an HTTP middleware that logs the request and response details.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
			headerWritten:  false,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.Log.Info("Request completed",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration.Milliseconds(),
			"size", responseData.size,
		)
	})
}
