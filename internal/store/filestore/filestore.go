package filestore

import (
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

func (fs *FileStore) Add(shortURL, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.URLMapping[shortURL] = originalURL

	fs.SaveToFile()
	logger.Log.Info("Added to store", "shortURL", shortURL, "originalURL", originalURL)

	return nil
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := fs.LoadFromFile(fs.FilePath)
	if err != nil {
		logger.Log.Error("Failed to load from file", "error", err)
		return "", err
	}

	originalURL, ok := fs.URLMapping[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	logger.Log.Info("Retrieved from store", "shortURL", shortURL, "originalURL", originalURL)

	return originalURL, nil
}

func (fs *FileStore) SaveToFile() error {
	file, err := os.Create("store.json")
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

		// Записываем объект ShortenStore в файл.
		if err := encoder.Encode(record); err != nil {
			return err
		}

		logger.Log.Info("Saving to file", "uuid", record.UUID, "short_url", shortURL, "original_url", originalURL)
	}

	return nil
}

func (fs *FileStore) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
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
