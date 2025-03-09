package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/contextutils"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
)

func init() {
	// Инициализация логгера для тестов
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

var (
	ErrURLNotFound = errors.New("URL not found")
)

// MockStore реализует интерфейс store.Store для тестирования
type MockStore struct {
	urls map[string]models.ShortenStore
}

func NewMockStore() *MockStore {
	return &MockStore{
		urls: make(map[string]models.ShortenStore),
	}
}

func (m *MockStore) Add(_ context.Context, shortURL, originalURL string, userID uuid.UUID) error {
	m.urls[shortURL] = models.ShortenStore{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	return nil
}

func (m *MockStore) Get(_ context.Context, shortURL string) (models.ShortenStore, error) {
	if record, ok := m.urls[shortURL]; ok {
		return record, nil
	}
	return models.ShortenStore{}, ErrURLNotFound
}

func (m *MockStore) AddBatch(_ context.Context, urls []models.ShortenBatchStore, userID uuid.UUID) error {
	for _, url := range urls {
		m.urls[url.ShortURL] = models.ShortenStore{
			UUID:        uuid.New(),
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
			UserID:      userID,
		}
	}
	return nil
}

func (m *MockStore) GetUserURLs(_ context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	var result []models.UserURLResponse
	for _, record := range m.urls {
		if record.UserID == userID {
			result = append(result, models.UserURLResponse{
				ShortURL:    record.ShortURL,
				OriginalURL: record.OriginalURL,
			})
		}
	}
	return result, nil
}

func (m *MockStore) DeleteUserURLs(_ context.Context, urls <-chan models.UserShortURL) error {
	for url := range urls {
		if record, ok := m.urls[url.ShortURL]; ok && record.UserID == url.UserID {
			record.Deleted = true
			m.urls[url.ShortURL] = record
		}
	}
	return nil
}

func (m *MockStore) Ping() error {
	return nil
}

// MockShortener реализует интерфейс services.Shortener для тестирования
type MockShortener struct{}

func (m *MockShortener) GenerateShortURL(url string) (string, error) {
	if url == "" {
		return "", ErrURLNotFound
	}
	return "test" + url[:5], nil
}

// addAuthCookie добавляет JWT токен в куки запроса
func addAuthCookie(req *http.Request) {
	userID := uuid.New()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("qwerty"))

	req.AddCookie(&http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
	})

	// Добавляем userID в контекст запроса
	ctx := contextutils.WithUserID(req.Context(), userID)
	*req = *req.WithContext(ctx)
}

func TestRouter_CreateShortLink(t *testing.T) {
	router := NewRouter()
	store := NewMockStore()
	shortener := &MockShortener{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	err := router.Routes(cfg, store, shortener)
	require.NoError(t, err)

	tests := []struct {
		name           string
		inputURL       string
		expectedStatus int
	}{
		{
			name:           "Valid URL",
			inputURL:       "https://example.com",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty URL",
			inputURL:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.inputURL))
			addAuthCookie(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestRouter_ShortenLink(t *testing.T) {
	router := NewRouter()
	store := NewMockStore()
	shortener := &MockShortener{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	err := router.Routes(cfg, store, shortener)
	require.NoError(t, err)

	tests := []struct {
		name           string
		request        models.ShortenRequest
		expectedStatus int
	}{
		{
			name: "Valid request",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Empty URL",
			request: models.ShortenRequest{
				URL: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
			addAuthCookie(req)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.ShortenResponse
				err = json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Result)
			}
		})
	}
}

func TestRouter_GetOriginalURL(t *testing.T) {
	router := NewRouter()
	store := NewMockStore()
	shortener := &MockShortener{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	err := router.Routes(cfg, store, shortener)
	require.NoError(t, err)

	// Добавляем тестовый URL в хранилище
	testShortURL := "testurl"
	testOriginalURL := "https://example.com"
	testUserID := uuid.New()
	err = store.Add(context.Background(), testShortURL, testOriginalURL, testUserID)
	require.NoError(t, err)

	tests := []struct {
		name           string
		shortURL       string
		expectedStatus int
	}{
		{
			name:           "Existing URL",
			shortURL:       testShortURL,
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "Non-existing URL",
			shortURL:       "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			addAuthCookie(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, testOriginalURL, w.Header().Get("Location"))
			}
		})
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	router := NewRouter()
	store := NewMockStore()
	shortener := &MockShortener{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	err := router.Routes(cfg, store, shortener)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/", nil)
	addAuthCookie(req)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
