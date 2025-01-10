package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
)

// ErrURLNotFound is an error that indicates the URL was not found.
var ErrURLNotFound = errors.New("URL not found")

// FileStore is a struct that represents the file store.
type FileStore struct {
	URLMapping map[string]string
	mu         sync.RWMutex
	FilePath   string
}

// Add is a method that adds a new URL to the file store.
func (fs *FileStore) Add(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.URLMapping[shortURL] = originalURL

	if fs.FilePath != "" {
		fs.SaveToFile()
	}

	logger.Log.Info("Added to store", "shortURL", shortURL, "originalURL", originalURL)

	return nil
}

// Get is a method that retrieves the original URL from the file store.
func (fs *FileStore) Get(ctx context.Context, shortURL string) (models.ShortenStore, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if fs.FilePath != "" {
		if err := fs.LoadFromFile(); err != nil {
			logger.Log.Error("Failed to load from file", "error", err)
			return models.ShortenStore{}, err
		}
	}

	originalURL, ok := fs.URLMapping[shortURL]
	if !ok {
		return models.ShortenStore{}, ErrURLNotFound
	}

	logger.Log.Info("Retrieved from store", "shortURL", shortURL, "originalURL", originalURL)

	return models.ShortenStore{
		OriginalURL: originalURL,
		Deleted:     false,
	}, nil
}

// AddBatch is a method that adds a batch of URLs to the file store.
func (fs *FileStore) AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, request := range batchRequest {
		fs.URLMapping[request.ShortURL] = request.OriginalURL
	}

	if fs.FilePath != "" {
		fs.SaveToFile()
	}

	logger.Log.Info("Added batch to store", "batchRequest", batchRequest)

	return nil
}

// SaveToFile is a method that saves the URL mapping to a file.
func (fs *FileStore) SaveToFile() error {
	file, err := os.OpenFile(fs.FilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for shortURL, originalURL := range fs.URLMapping {
		record := models.ShortenStore{
			UUID:        uuid.New(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		}

		if err := encoder.Encode(&record); err != nil {
			return err
		}

		logger.Log.Info("Saving to file", "uuid", record.UUID, "short_url", shortURL, "original_url", originalURL)
	}

	return nil
}

// LoadFromFile is a method that loads the URL mapping from a file.
func (fs *FileStore) LoadFromFile() error {
	file, err := os.Open(fs.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	for {
		var record models.ShortenStore
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fs.URLMapping[record.ShortURL] = record.OriginalURL
	}

	logger.Log.Info("Loaded from file", "URLMapping", fs.URLMapping)
	return nil
}

// GetUserURLs is a method that retrieves all URLs associated with the user ID.
func (fs *FileStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	return nil, nil
}

// DeleteUserURLs is a method that deletes URLs associated with the user ID.
func (fs *FileStore) DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
	return nil
}

// Ping is a method that checks the file store connection.
func (fs *FileStore) Ping() error {
	err := errors.New("unable to access the store")
	return err
}
