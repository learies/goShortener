package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/learies/goShortener/internal/config/contextutils"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	// Mock the next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := contextutils.GetUserID(r.Context())
		if ok {
			w.Write([]byte("UserID: " + userID.String()))
		} else {
			w.Write([]byte("No userID"))
		}
	})

	tests := []struct {
		name           string
		setupRequest   func(req *http.Request)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid Token Provided",
			setupRequest: func(req *http.Request) {
				claims := &Claims{
					UserID: "12345678-1234-1234-1234-123456789abc",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("qwerty"))
				req.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "UserID: 12345678-1234-1234-1234-123456789abc",
		},
		{
			name: "No Token Provided",
			setupRequest: func(req *http.Request) {
				// No token to setup
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "UserID: ", // In reality, this will be a new UUID, verify format instead
		},
		{
			name: "Invalid Token Provided",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to the middleware
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			// Set up the request based on the test case
			tt.setupRequest(req)

			// Create a ResponseRecorder to record the response
			rec := httptest.NewRecorder()

			// Run the middleware with the mock handler
			handler := JWTMiddleware(nextHandler)
			handler.ServeHTTP(rec, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check the response body
			responseBody, err := io.ReadAll(rec.Body)
			assert.NoError(t, err)
			assert.True(t, strings.HasPrefix(string(responseBody), tt.expectedBody))
		})
	}
}
