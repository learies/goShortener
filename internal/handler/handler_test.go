package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(originalURL string) string {
	return "short123"
}

type MockStore struct{}

func (m *MockStore) Add(shortURL, originalURL string) error {
	return nil
}

func (m *MockStore) Get(shortURL string) (string, error) {
	return "http://example.com", nil
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

		if result.StatusCode != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", result.Status)
		}

		body, err := io.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expected := "http://localhost/short123"
		if string(body) != expected {
			t.Errorf("Expected body %q, got %q", expected, string(body))
		}
	})

	t.Run("GetOriginalURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/short123", nil)
		recorder := httptest.NewRecorder()

		handler.GetOriginalURL(mockStore)(recorder, req)

		result := recorder.Result()

		if result.StatusCode != http.StatusTemporaryRedirect {
			t.Errorf("Expected status Temporary Redirect, got %v", result.Status)
		}

		body, err := io.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expected := "http://example.com"
		if string(body) != expected {
			t.Errorf("Expected body %q, got %q", expected, string(body))
		}
	})
}
