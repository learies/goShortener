package filestore

import (
	"errors"
	"sync"

	"github.com/learies/goShortener/internal/config/logger"
)

var ErrURLNotFound = errors.New("URL not found")

type FileStore struct {
	URLMapping map[string]string
	mu         sync.Mutex
}

func (fs *FileStore) Add(shortURL, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.URLMapping[shortURL] = originalURL

	logger.Log.Info("Added to store", "shortURL", shortURL, "originalURL", originalURL)

	return nil
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	originalURL, ok := fs.URLMapping[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	logger.Log.Info("Retrieved from store", "shortURL", shortURL, "originalURL", originalURL)

	return originalURL, nil
}
