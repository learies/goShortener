package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(originalURL string) string {
	return "short123"
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
		reqBody := strings.NewReader("http://example.com")
		req := httptest.NewRequest(http.MethodPost, "/", reqBody)
		recorder := httptest.NewRecorder()

		handler.CreateShortLink(mockStore, "http://localhost", mockShortener)(recorder, req)

		result := recorder.Result()
		assert.Equal(t, http.StatusCreated, result.StatusCode)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := "http://localhost/short123"
		assert.Equal(t, expected, string(body))
	})

	t.Run("GetOriginalURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/short123", nil)
		recorder := httptest.NewRecorder()

		mockStore.GetFunc = func(shortURL string) (string, error) {
			return "http://example.com", nil
		}

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()

		assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode)

		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)

		expected := "http://example.com"
		assert.Equal(t, expected, string(body))
	})
}
