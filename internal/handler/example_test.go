package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"

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

func ExampleHandler_CreateShortLink() {
	h := handler.NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}
	baseURL := "http://localhost:8080"

	reqBody := strings.NewReader("https://example.com/")
	req := httptest.NewRequest("POST", "/", reqBody)
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()

	userID := uuid.New()
	ctx := contextutils.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	h.CreateShortLink(mockStore, baseURL, mockShortener)(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: http://localhost:8080/EwHXdJfB
}

func ExampleHandler_CreateShortLink_badRequest() {
	h := handler.NewHandler()
	mockStore := &MockStore{}
	mockShortener := &MockShortener{}
	baseURL := "http://localhost:8080"

	reqBody := strings.NewReader("{ bad json")
	req := httptest.NewRequest("POST", "/", reqBody)
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()

	userID := uuid.New()
	ctx := contextutils.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	h.CreateShortLink(mockStore, baseURL, mockShortener)(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Println(res.Status)
	// Output: 400 Bad Request
}

func ExampleHandler_GetOriginalURL_notFound() {
	h := handler.NewHandler()
	mockStore := &MockStore{}
	mockStore.GetFunc = func(ctx context.Context, shortURL string) (models.ShortenStore, error) {
		return models.ShortenStore{}, filestore.ErrURLNotFound
	}

	req := httptest.NewRequest("GET", "/EwHXdJfB", nil)
	rec := httptest.NewRecorder()

	h.GetOriginalURL(mockStore)(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: URL not found
}
