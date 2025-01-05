package store

import (
	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/store/filestore"
)

type Store interface {
	Add(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

func NewStore(cfg config.Config) (Store, error) {
	store := &filestore.FileStore{
		URLMapping: make(map[string]string),
		FilePath:   cfg.FilePath,
	}

	return store, nil
}
