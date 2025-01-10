package models

import (
	"github.com/google/uuid"
)

// ShortenRequest is a struct that represents the request body for shortening a URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is a struct that represents the response body for shortening a URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenStore is a struct that represents the data stored for a shortened URL.
type ShortenStore struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      uuid.UUID `json:"user_id"`
	Deleted     bool      `json:"deleted"`
}

// ShortenBatchRequest is a struct that represents the request body for batch shortening URLs.
type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResponse is a struct that represents the response body for batch shortening URLs.
type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLResponse is a struct that represents the response body for a user's URL.
type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ShortenBatchStore is a struct that represents the data stored for a batch of shortened URLs.
type ShortenBatchStore struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
}

// ShortenDeleteRequest is a struct that represents the request body for deleting URLs.
type ShortenDeleteRequest struct {
	ShortURLs []string `json:"short_urls"`
}

// UserShortURL is a struct that represents the data stored for a user's short URL.
type UserShortURL struct {
	UserID   uuid.UUID `json:"user_id"`
	ShortURL string    `json:"short_url"`
}
