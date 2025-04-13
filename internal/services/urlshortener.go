package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store"
)

// URLShortenerService provides business logic for URL shortening operations
type URLShortenerService struct {
	store   store.Store
	baseURL string
}

// NewURLShortenerService creates a new URLShortenerService instance
func NewURLShortenerService(store store.Store, baseURL string) *URLShortenerService {
	return &URLShortenerService{
		store:   store,
		baseURL: baseURL,
	}
}

// GenerateShortURL implements the Shortener interface
func (s *URLShortenerService) GenerateShortURL(urlStr string) (string, error) {
	if urlStr == "" {
		return "", ErrEmptyURL
	}

	if _, err := url.Parse(urlStr); err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	shortURL := uuid.New().String()[:8]
	return shortURL, nil
}

// CreateShortURL creates a short URL for the given original URL
func (s *URLShortenerService) CreateShortURL(ctx context.Context, originalURL string, userID uuid.UUID) (string, error) {
	shortURL, err := s.GenerateShortURL(originalURL)
	if err != nil {
		return "", err
	}

	if err := s.store.Add(ctx, shortURL, originalURL, userID); err != nil {
		return "", fmt.Errorf("failed to store URL: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, shortURL), nil
}

// GetOriginalURL retrieves the original URL for a given short URL
func (s *URLShortenerService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	store, err := s.store.Get(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	if store.Deleted {
		return "", errors.New("URL has been deleted")
	}

	return store.OriginalURL, nil
}

// CreateBatchShortURL creates multiple short URLs in batch
func (s *URLShortenerService) CreateBatchShortURL(ctx context.Context, batchRequest []models.ShortenBatchRequest, userID uuid.UUID) ([]models.ShortenBatchResponse, error) {
	batchStore := make([]models.ShortenBatchStore, len(batchRequest))
	for i, req := range batchRequest {
		shortURL, err := s.GenerateShortURL(req.OriginalURL)
		if err != nil {
			return nil, err
		}
		batchStore[i] = models.ShortenBatchStore{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
			ShortURL:      shortURL,
		}
	}

	if err := s.store.AddBatch(ctx, batchStore, userID); err != nil {
		return nil, fmt.Errorf("failed to store batch URLs: %w", err)
	}

	response := make([]models.ShortenBatchResponse, len(batchStore))
	for i, store := range batchStore {
		response[i] = models.ShortenBatchResponse{
			CorrelationID: store.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, store.ShortURL),
		}
	}

	return response, nil
}

// GetUserURLs retrieves all URLs created by a user
func (s *URLShortenerService) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	urls, err := s.store.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	response := make([]models.UserURLResponse, len(urls))
	for i, url := range urls {
		response[i] = models.UserURLResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, url.ShortURL),
			OriginalURL: url.OriginalURL,
		}
	}

	return response, nil
}

// DeleteUserURLs deletes URLs created by a user
func (s *URLShortenerService) DeleteUserURLs(ctx context.Context, userID uuid.UUID, shortURLs []string) error {
	urlChan := make(chan models.UserShortURL, len(shortURLs))
	for _, shortURL := range shortURLs {
		urlChan <- models.UserShortURL{
			UserID:   userID,
			ShortURL: shortURL,
		}
	}
	close(urlChan)

	if err := s.store.DeleteUserURLs(ctx, urlChan); err != nil {
		return fmt.Errorf("failed to delete user URLs: %w", err)
	}

	return nil
}

// GetStats retrieves service statistics
func (s *URLShortenerService) GetStats(ctx context.Context) (int, int, error) {
	return s.store.GetStats(ctx)
}
