package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/store"
)

// StatsResponse represents the response structure for the stats endpoint
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// GetStats is a handler that returns statistics about the URL shortener service
func GetStats(store store.Store, trustedSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if trusted subnet is configured
		if trustedSubnet == "" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Get client IP from X-Real-IP header
		clientIP := r.Header.Get("X-Real-IP")
		if clientIP == "" {
			http.Error(w, "Missing X-Real-IP header", http.StatusForbidden)
			return
		}

		// Parse the trusted subnet
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Log.Error("Failed to parse trusted subnet", "error", err)
			http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
			return
		}

		// Parse client IP
		ip := net.ParseIP(clientIP)
		if ip == nil {
			http.Error(w, "Invalid client IP", http.StatusForbidden)
			return
		}

		// Check if client IP is in trusted subnet
		if !ipNet.Contains(ip) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Get stats from store
		urlsCount, usersCount, err := store.GetStats(r.Context())
		if err != nil {
			logger.Log.Error("Failed to get stats", "error", err)
			http.Error(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}

		// Prepare response
		response := StatsResponse{
			URLs:  urlsCount,
			Users: usersCount,
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Encode and send response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Log.Error("Failed to encode response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
