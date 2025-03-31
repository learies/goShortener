// Package handler provides HTTP request handlers for the URL shortener service.
package handler

import (
	"net/http"

	"github.com/learies/goShortener/internal/store"
)

// Handler is a struct that represents the handler.
type Handler struct {
}

// NewHandler is a function that creates a new handler.
func NewHandler() *Handler {
	return &Handler{}
}

// GetStats is a method that returns statistics about the URL shortener service
func (h *Handler) GetStats(store store.Store, trustedSubnet string) http.HandlerFunc {
	return GetStats(store, trustedSubnet)
}
