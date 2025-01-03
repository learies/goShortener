package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/learies/goShortener/internal/store/filestore"
)

type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(originalURL string) string {
	return "EwHXdJfB"
}

type MockStore struct {
	GetFunc func(shortURL string) (string, error)
}

func (m *MockStore) Add(shortURL, originalURL string) error {
	return nil
}

func (m *MockStore) Get(shortURL string) (string, error) {
	return m.GetFunc(shortURL)
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

		handler.CreateShortLink(mockStore, "http://localhost:8080", mockShortener)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("GetOriginalURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(shortURL string) (string, error) {
			return "https://practicum.yandex.ru/", nil
		}

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode)
		assert.Equal(t, "https://practicum.yandex.ru/", result.Header.Get("Location"))
	})

	t.Run("GetOriginalURLNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(shortURL string) (string, error) {
			return "", filestore.ErrURLNotFound
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
}
