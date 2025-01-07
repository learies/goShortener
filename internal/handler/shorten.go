package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

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

		shortenedURL := baseURL + "/" + shortURL

		err = store.Add(ctx, shortURL, originalURL)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortenedURL))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}

func (h *Handler) GetOriginalURL(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		originalURL, err := store.Get(ctx, shortURL)
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
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

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

		shortenedURL := baseURL + "/" + shortURL

		var shortenResponse models.ShortenResponse
		shortenResponse.Result = shortenedURL

		responseBody, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, "can't marshal response", http.StatusInternalServerError)
			return
		}

		err = store.Add(ctx, shortURL, originalURL)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(responseBody)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseBody)
	}
}

func (h *Handler) ShortenLinkBatch(store store.Store, baseURL string, shortener services.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}

		var batchRequest []models.ShortenBatchRequest
		err = json.Unmarshal(body, &batchRequest)
		if err != nil {
			http.Error(w, "can't unmarshal body", http.StatusBadRequest)
			return
		}

		var batchResponse []models.ShortenBatchResponse
		var batchShorten []models.ShortenBatchStore
		for _, request := range batchRequest {
			shortURL, err := shortener.GenerateShortURL(request.OriginalURL)
			if err != nil {
				http.Error(w, "can't generate short URL", http.StatusInternalServerError)
				return
			}

			shortenedURL := baseURL + "/" + shortURL

			batchResponse = append(batchResponse, models.ShortenBatchResponse{
				CorrelationID: request.CorrelationID,
				ShortURL:      shortenedURL,
			})

			batchShorten = append(batchShorten, models.ShortenBatchStore{
				CorrelationID: request.CorrelationID,
				ShortURL:      shortURL,
				OriginalURL:   request.OriginalURL,
			})
		}

		err = store.AddBatch(ctx, batchShorten)
		if err != nil {
			http.Error(w, "can't save batch short URL", http.StatusInternalServerError)
			return
		}

		responseBody, err := json.Marshal(batchResponse)
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
