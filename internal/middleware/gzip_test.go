package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware_ResponseWithGzip(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler := GzipMiddleware(next)
	handler.ServeHTTP(rr, req)

	if encoding := rr.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %v", encoding)
	}

	gr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	body, _ := io.ReadAll(gr)
	expectedBody := "Hello, World!"
	if string(body) != expectedBody {
		t.Errorf("expected body %v, got %v", expectedBody, string(body))
	}
}

func TestGzipMiddleware_ResponseWithoutGzip(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("Accept-Encoding", "deflate")

	rr := httptest.NewRecorder()
	handler := GzipMiddleware(next)
	handler.ServeHTTP(rr, req)

	if encoding := rr.Header().Get("Content-Encoding"); encoding != "" {
		t.Errorf("expected Content-Encoding , got %v", encoding)
	}

	expectedBody := "Hello, World!"
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("expected body %v, got %v", expectedBody, strings.TrimSpace(rr.Body.String()))
	}
}

func TestGzipMiddleware_InvalidGzipRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer([]byte("invalidgzip")))
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler := GzipMiddleware(next)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %v, got %v", http.StatusBadRequest, rr.Code)
	}

	expectedBody := "Invalid gzip content\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %v, got %v", expectedBody, rr.Body.String())
	}
}

func TestGzipMiddleware_ValidGzipRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		expectedBody := "Hello, World!"
		if string(body) != expectedBody {
			t.Errorf("expected body %v, got %v", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received"))
	})

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("Hello, World!"))
	gz.Close()

	req := httptest.NewRequest(http.MethodPost, "http://example.com", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler := GzipMiddleware(next)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	expectedResponse := "Received"
	if rr.Body.String() != expectedResponse {
		t.Errorf("expected body %v, got %v", expectedResponse, rr.Body.String())
	}
}
