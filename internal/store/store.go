package store

import (
	"context"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/database"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store/dbstore"
	"github.com/learies/goShortener/internal/store/filestore"
)

type Store interface {
	Add(ctx context.Context, shortURL, originalURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
	AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore) error
	Ping() error
}

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
