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

func TestMainHandler(t *testing.T) {
	h := NewHandler()

	mockShortener := &MockShortener{}

	t.Run("CreateShortLink", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("http://example.com"))
		w := httptest.NewRecorder()

		h.CreateShortLink("http://localhost", mockShortener)(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", res.Status)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Couldn't read response body: %v", err)
		}

		expected := "http://localhost/short123"
		if string(body) != expected {
			t.Errorf("Expected body %v, got %v", expected, string(body))
		}
	})
}
