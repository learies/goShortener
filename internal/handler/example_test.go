package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/learies/goShortener/internal/config/contextutils"
	"github.com/learies/goShortener/internal/handler"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store/filestore"
)

type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(originalURL string) (string, error) {
	return "EwHXdJfB", nil
}

type MockStore struct {
	handler.MockStore
}

func TestHandler(t *testing.T) {
	h := handler.NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}
	baseURL := "http://localhost:8080"

	t.Run("CreateShortLink", func(t *testing.T) {
		reqBody := strings.NewReader("https://example.com/")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		h.CreateShortLink(mockStore, baseURL, mockShortener)(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		contentType := res.Header.Get("Content-Type")
		assert.Equal(t, "text/plain", contentType)

		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		expected := "http://localhost:8080/EwHXdJfB"
		assert.Equal(t, expected, string(body))
	})

	t.Run("CreateShortLinkBadRequest", func(t *testing.T) {
		reqBody := strings.NewReader("{ bad json")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		userID := uuid.New()
		ctx := contextutils.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		h.CreateShortLink(mockStore, baseURL, mockShortener)(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	// Additional test cases would be similarly implemented...

	t.Run("GetOriginalURLNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		rec := httptest.NewRecorder()

		mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
			return models.ShortenStore{}, filestore.ErrURLNotFound
		}

		h.GetOriginalURL(mockStore)(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		expected := "URL not found\n"
		assert.Equal(t, expected, string(body))
	})

	// Add similar blocks for other test cases
}
