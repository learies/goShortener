package store

import (
	"github.com/learies/goShortener/internal/store/filestore"
)

type Store interface {
	Add(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

func NewStore() (Store, error) {
	store := &filestore.FileStore{
		URLMapping: make(map[string]string),
	}

	return store, nil
}
