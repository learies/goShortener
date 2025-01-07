package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
)

var ErrURLNotFound = errors.New("URL not found")

type FileStore struct {
	URLMapping map[string]string
	mu         sync.Mutex
	FilePath   string
}

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

func (fs *FileStore) Get(ctx context.Context, shortURL string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.FilePath != "" {
		err := fs.LoadFromFile()
		if err != nil {
			logger.Log.Error("Failed to load from file", "error", err)
			return "", err
		}
	}

	originalURL, ok := fs.URLMapping[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	logger.Log.Info("Retrieved from store", "shortURL", shortURL, "originalURL", originalURL)

	return originalURL, nil
}

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
			break
		}

		fs.URLMapping[record.ShortURL] = record.OriginalURL
	}

	logger.Log.Info("Loaded from file", "URLMapping", fs.URLMapping)

	return nil
}

func (fs *FileStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	return nil, nil
}

func (fs *FileStore) DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
	return nil
}

func (fs *FileStore) Ping() error {
	err := errors.New("unable to access the store")
	return err
}
