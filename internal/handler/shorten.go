package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/learies/goShortener/internal/config/contextutils"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/services"
	"github.com/learies/goShortener/internal/services/worker"
	"github.com/learies/goShortener/internal/store"
)

// checkOriginalURL verifies if the provided URL starts with the prefixes
// "http://" or "https://". It returns true if the URL is valid, otherwise false.
func checkOriginalURL(originalURL string) bool {
	return strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://")
}

// CreateShortLink is an HTTP handler that reads an original URL from the request
// body, generates a short URL, and responds with the shortened URL.
// It requires a store to persist the mapping and a shortener to generate the short URL.
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

		userID, ok := contextutils.GetUserID(ctx)
		if !ok {
			http.Error(w, "UserID not found in context", http.StatusUnauthorized)
			return
		}

		err = store.Add(ctx, shortURL, originalURL, userID)
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

// GetOriginalURL is an HTTP handler that retrieves the original URL for a given
// short URL path and redirects the client.
// It requires a store to fetch the mapping from the short URL.
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

		if originalURL.Deleted {
			http.Error(w, "URL is deleted", http.StatusGone)
			return
		}

		w.Header().Set("Location", originalURL.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// ShortenLink is an HTTP handler that reads a JSON body with an original URL,
// generates a short URL, and responds with a JSON containing the shortened URL.
// It requires a store to persist the mapping and a shortener to generate the short URL.
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

		userID, ok := contextutils.GetUserID(ctx)
		if !ok {
			http.Error(w, "UserID not found in context", http.StatusUnauthorized)
			return
		}

		err = store.Add(ctx, shortURL, originalURL, userID)
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

// ShortenLinkBatch is an HTTP handler that reads a JSON array of URLs,
// generates short URLs for each, and responds with a JSON array of shortened URLs.
// It requires a store to persist the batch and a shortener to generate short URLs.
func (h *Handler) ShortenLinkBatch(store store.Store, baseURL string, shortener services.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		userID, ok := contextutils.GetUserID(ctx)
		if !ok {
			http.Error(w, "UserID not found in context", http.StatusUnauthorized)
			return
		}

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

		err = store.AddBatch(ctx, batchShorten, userID)
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

// PingHandler is an HTTP handler that checks the availability of the store.
// It responds with a success message if the store is reachable.
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

// GetUserURLs is an HTTP handler that retrieves all URLs associated with the user
// ID in the context. It responds with a JSON array of these URLs.
// It requires a store to fetch the user's URLs.
func (h *Handler) GetUserURLs(store store.Store, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		userID, ok := contextutils.GetUserID(ctx)
		if !ok {
			http.Error(w, "UserID not found in context", http.StatusUnauthorized)
			return
		}

		urls, err := store.GetUserURLs(ctx, userID)
		if err != nil {
			http.Error(w, "can't get user URLs", http.StatusNotFound)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		modifiedUrls := make([]models.UserURLResponse, len(urls))
		for i, url := range urls {
			modifiedUrls[i] = models.UserURLResponse{
				ShortURL:    baseURL + "/" + url.ShortURL,
				OriginalURL: url.OriginalURL,
			}
		}

		responseBody, err := json.Marshal(modifiedUrls)
		if err != nil {
			http.Error(w, "can't marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}
}

// DeleteUserURLs is an HTTP handler that reads a JSON array of short URLs to be
// deleted for the user. It performs logical deletion of these URLs.
// It requires a store to delete the URLs and the user's ID in the context.
func (h *Handler) DeleteUserURLs(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		userID, ok := contextutils.GetUserID(ctx)
		if !ok {
			http.Error(w, "UserID not found in context", http.StatusUnauthorized)
			return
		}

		var deleteRequest models.ShortenDeleteRequest
		err := json.NewDecoder(r.Body).Decode(&deleteRequest.ShortURLs)
		if err != nil {
			http.Error(w, "can't unmarshal body", http.StatusBadRequest)
			return
		}

		for _, shortURL := range deleteRequest.ShortURLs {
			err := store.DeleteUserURLs(ctx, worker.DeleteUserURLs(models.UserShortURL{
				UserID:   userID,
				ShortURL: shortURL,
			}))
			if err != nil {
				http.Error(w, "can't delete URL", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
