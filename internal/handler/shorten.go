package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/learies/goShortener/internal/services"
)

func checkOriginalURL(originalURL string) bool {
	return strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://")
}

func (h *Handler) CreateShortLink(baseURL string, shortener services.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}

		originalURL := string(body)
		if !checkOriginalURL(originalURL) {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		shortURL := shortener.GenerateShortURL(originalURL)
		shortenedURL := baseURL + "/" + shortURL

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}
