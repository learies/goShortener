package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/learies/goShortener/internal/config/contextutils"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store/filestore"
)

func init() {
	// Инициализация логгера для тестов
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(originalURL string) (string, error) {
	return "EwHXdJfB", nil
}

type MockStore struct {
	GetFunc            func(ctx context.Context, shortURL string) (models.ShortenStore, error)
	AddFunc            func(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error
	AddBatchFunc       func(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error
	GetUserURLsFunc    func(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error)
	DeleteUserURLsFunc func(ctx context.Context, userShortURLs <-chan models.UserShortURL) error
	PingFunc           func() error
}

func (m *MockStore) Add(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
	if m.AddFunc != nil {
		return m.AddFunc(ctx, shortURL, originalURL, userID)
	}
	return nil
}

func (m *MockStore) Get(ctx context.Context, shortURL string) (models.ShortenStore, error) {
	return m.GetFunc(ctx, shortURL)
}

func (m *MockStore) AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error {
	if m.AddBatchFunc != nil {
		return m.AddBatchFunc(ctx, batchRequest, userID)
	}
	return nil
}

func (m *MockStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	if m.GetUserURLsFunc != nil {
		return m.GetUserURLsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
	return nil
}

func (m *MockStore) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

func TestMainHandler(t *testing.T) {
	handler := NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}

	t.Run("CreateShortLink", func(t *testing.T) {
		reqBody := strings.NewReader("https://practicum.yandex.ru/")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		req.Header.Set("Content-Type", "text/plain")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.CreateShortLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()
		assert.Equal(t, http.StatusCreated, result.StatusCode)

		contentType := result.Header.Get("Content-Type")
		assert.Equal(t, "text/plain", contentType)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := "http://localhost:8080/EwHXdJfB"
		assert.Equal(t, expected, string(body))
	})

	t.Run("CreateShortLinkBadRequest", func(t *testing.T) {
		reqBody := strings.NewReader("{ bad json")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		req.Header.Set("Content-Type", "text/plain")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.CreateShortLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("CreateShortLinkConflict", func(t *testing.T) {
		reqBody := strings.NewReader("https://practicum.yandex.ru/")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		req.Header.Set("Content-Type", "text/plain")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.AddFunc = func(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
			return fmt.Errorf("conflict error")
		}

		handler.CreateShortLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusConflict, result.StatusCode)
	})

	t.Run("GetOriginalURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
			return models.ShortenStore{
				OriginalURL: "https://practicum.yandex.ru/",
				Deleted:     false,
			}, nil
		}

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode)
		assert.Equal(t, "https://practicum.yandex.ru/", result.Header.Get("Location"))
	})

	t.Run("GetDeletedOriginalURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
			return models.ShortenStore{Deleted: true}, nil
		}

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusGone, result.StatusCode)
	})

	t.Run("GetOriginalURLNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
			return models.ShortenStore{}, filestore.ErrURLNotFound
		}

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusNotFound, result.StatusCode)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := "URL not found\n"
		assert.Equal(t, expected, string(body))
	})

	t.Run("ShortenLink", func(t *testing.T) {
		reqBody := `{"url":"https://practicum.yandex.ru/"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.AddFunc = func(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
			return nil
		}

		handler.ShortenLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusCreated, result.StatusCode)

		contentType := result.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := `{"result":"http://localhost:8080/EwHXdJfB"}`
		assert.JSONEq(t, expected, string(body))
	})

	t.Run("ShortenLinkConflict", func(t *testing.T) {
		reqBody := `{"url":"https://practicum.yandex.ru/"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.AddFunc = func(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
			return fmt.Errorf("conflict error")
		}

		handler.ShortenLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusConflict, result.StatusCode)
	})

	t.Run("ShortenLinkBadRequest", func(t *testing.T) {
		reqBody := `{"bad json"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.ShortenLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("ShortenLinkBatch", func(t *testing.T) {
		reqBody := `[
			{"correlation_id": "1", "original_url": "https://practicum.yandex.ru/"},
			{"correlation_id": "2", "original_url": "https://yandex.ru/"}
		]`
		req := httptest.NewRequest(http.MethodPost, "/shorten/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.AddBatchFunc = func(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error {
			return nil
		}

		handler.ShortenLinkBatch(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusCreated, result.StatusCode)

		contentType := result.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := `[
			{"correlation_id": "1", "short_url": "http://localhost:8080/EwHXdJfB"},
			{"correlation_id": "2", "short_url": "http://localhost:8080/EwHXdJfB"}
		]`
		assert.JSONEq(t, expected, string(body))
	})

	t.Run("ShortenLinkBatchBadRequest", func(t *testing.T) {
		reqBody := `[{ bad json }]`
		req := httptest.NewRequest(http.MethodPost, "/shorten/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.ShortenLinkBatch(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.GetUserURLsFunc = func(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
			return []models.UserURLResponse{
				{
					ShortURL:    "EwHXdJfB",
					OriginalURL: "https://practicum.yandex.ru/",
				},
			}, nil
		}

		handler.GetUserURLs(mockStore, "http://localhost:8080")(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)

		contentType := result.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := `[
			{"short_url": "http://localhost:8080/EwHXdJfB", "original_url": "https://practicum.yandex.ru/"}
		]`
		assert.JSONEq(t, expected, string(body))
	})

	t.Run("GetUserURLsEmpty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.GetUserURLsFunc = func(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
			return []models.UserURLResponse{}, nil
		}

		handler.GetUserURLs(mockStore, "http://localhost:8080")(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusNoContent, result.StatusCode)
	})

	t.Run("DeleteUserURLs", func(t *testing.T) {
		reqBody := `["EwHXdJfB", "AbCdEfGh"]`
		req := httptest.NewRequest(http.MethodDelete, "/user/urls", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		mockStore.DeleteUserURLsFunc = func(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
			return nil
		}

		handler.DeleteUserURLs(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusAccepted, result.StatusCode)
	})

	t.Run("DeleteUserURLsBadRequest", func(t *testing.T) {
		reqBody := `{ bad json }`
		req := httptest.NewRequest(http.MethodDelete, "/user/urls", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.DeleteUserURLs(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("PingHandler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		recorder := httptest.NewRecorder()

		mockStore.PingFunc = func() error {
			return nil
		}

		handler.PingHandler(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := "Successfully connected to the store"
		assert.Equal(t, expected, string(body))
	})

	t.Run("PingHandlerError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		recorder := httptest.NewRecorder()

		mockStore.PingFunc = func() error {
			return fmt.Errorf("store error")
		}

		handler.PingHandler(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}

func BenchmarkCreateShortLink(b *testing.B) {
	handler := NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}
	baseURL := "http://localhost:8080"

	reqBody := []byte("https://practicum.yandex.ru/")

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		recorder := httptest.NewRecorder()
		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.CreateShortLink(mockStore, baseURL, mockShortener)(recorder, req)
	}
}

func BenchmarkGetOriginalURL(b *testing.B) {
	handler := NewHandler()
	mockStore := &MockStore{}

	mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
		return models.ShortenStore{
			OriginalURL: "https://practicum.yandex.ru/",
			Deleted:     false,
		}, nil
	}

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		handler.GetOriginalURL(mockStore)(recorder, req)
	}
}

func BenchmarkShortenLink(b *testing.B) {
	handler := NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}
	baseURL := "http://localhost:8080"

	reqBody := `{"url":"https://practicum.yandex.ru/"}`

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		recorder := httptest.NewRecorder()
		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		handler.ShortenLink(mockStore, baseURL, mockShortener)(recorder, req)
	}
}
