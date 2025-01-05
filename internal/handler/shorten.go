package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/services"
	"github.com/learies/goShortener/internal/store"
)

func checkOriginalURL(originalURL string) bool {
	return strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://")
}

func (h *Handler) CreateShortLink(store store.Store, baseURL string, shortener services.Shortener) http.HandlerFunc {
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

		shortURL, err := shortener.GenerateShortURL(originalURL)
		if err != nil {
			http.Error(w, "can't generate short URL", http.StatusInternalServerError)
			return
		}

		err = store.Add(shortURL, originalURL)
		if err != nil {
			http.Error(w, "can't save short URL", http.StatusInternalServerError)
			return
		}

		shortenedURL := baseURL + "/" + shortURL

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}

func (h *Handler) GetOriginalURL(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		originalURL, err := store.Get(shortURL)
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) ShortenLink(store store.Store, baseURL string, shortener services.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}

		var shortenRequest models.ShortenRequest
		err = json.Unmarshal(body, &shortenRequest)
		if err != nil {
			http.Error(w, "can't unmarshal body", http.StatusBadRequest)
			return
		}

		originalURL := string(shortenRequest.URL)
		if !checkOriginalURL(originalURL) {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		shortURL, err := shortener.GenerateShortURL(originalURL)
		if err != nil {
			http.Error(w, "can't generate short URL", http.StatusInternalServerError)
			return
		}

		err = store.Add(shortURL, originalURL)
		if err != nil {
			http.Error(w, "can't save short URL", http.StatusInternalServerError)
			return
		}

		shortenedURL := baseURL + "/" + shortURL

		var shortenResponse models.ShortenResponse
		shortenResponse.Result = shortenedURL

		responseBody, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, "can't marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseBody)
	}
}

func (h *Handler) PingHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			http.Error(w, "Store is not available", http.StatusInternalServerError)
			logger.Log.Error("Store ping failed", "error", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Successfully connected to the store"))
	}
}
