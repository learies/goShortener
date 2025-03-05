// Package store defines the storage interface and implementations for the URL shortener service.
package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/database"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store/dbstore"
	"github.com/learies/goShortener/internal/store/filestore"
)

// Store is an interface that defines the methods for the store.
type Store interface {
	Add(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error
	Get(ctx context.Context, shortURL string) (models.ShortenStore, error)
	AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error)
	DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error
	Ping() error
}

// NewStore is a function that creates a new store.
func NewStore(cfg config.Config) (Store, error) {
	if cfg.DatabaseDSN != "" {
		db, err := database.Connect(cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return &dbstore.DBStore{DB: db}, nil
	}

	store := &filestore.FileStore{
		URLMapping: make(map[string]string),
		FilePath:   cfg.FilePath,
	}

	return store, nil
}
