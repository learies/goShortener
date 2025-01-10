package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.NewLogger("debug")
}

// mockHandler is a simple HTTP handler to be used for testing.
func mockHandler(statusCode int, responseBody string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	})
}

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedStatus int
		expectedSize   int
	}{
		{
			name:           "Successful Response",
			statusCode:     http.StatusOK,
			responseBody:   "Hello, World!",
			expectedStatus: http.StatusOK,
			expectedSize:   len("Hello, World!"),
		},
		{
			name:           "Not Found Response",
			statusCode:     http.StatusNotFound,
			responseBody:   "Not Found",
			expectedStatus: http.StatusNotFound,
			expectedSize:   len("Not Found"),
		},
		{
			name:           "Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   "Internal Error",
			expectedStatus: http.StatusInternalServerError,
			expectedSize:   len("Internal Error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP request and response recorder
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			rec := httptest.NewRecorder()
			handlerWithLogging := WithLogging(mockHandler(tc.statusCode, tc.responseBody))

			// Serve the HTTP request using the middleware
			handlerWithLogging.ServeHTTP(rec, req)

			// Verify the status code and response size
			assert.Equal(t, tc.expectedStatus, rec.Code)
			response := rec.Body.String()
			assert.Equal(t, tc.responseBody, response)
			assert.Equal(t, tc.expectedSize, rec.Body.Len())
		})
	}
}
