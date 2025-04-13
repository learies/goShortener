package filestore

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
)

func init() {
	// Инициализация логгера для тестов
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestFileStore(t *testing.T) {
	// Создаем временную директорию для тестовых файлов
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "urls.json")

	// Создаем новый экземпляр FileStore
	fs := &FileStore{
		URLMapping: make(map[string]string),
		FilePath:   filePath,
	}

	// Тестовые данные
	shortURL := "abc123"
	originalURL := "https://example.com"
	userID := uuid.New()

	t.Run("Add and Get", func(t *testing.T) {
		// Добавляем URL
		err := fs.Add(context.Background(), shortURL, originalURL, userID)
		require.NoError(t, err)

		// Получаем URL
		result, err := fs.Get(context.Background(), shortURL)
		require.NoError(t, err)
		assert.Equal(t, originalURL, result.OriginalURL)
		assert.False(t, result.Deleted)

		// Проверяем несуществующий URL
		_, err = fs.Get(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, ErrURLNotFound)
	})

	t.Run("AddBatch", func(t *testing.T) {
		batchRequest := []models.ShortenBatchStore{
			{
				CorrelationID: "1",
				ShortURL:      "short1",
				OriginalURL:   "https://example1.com",
			},
			{
				CorrelationID: "2",
				ShortURL:      "short2",
				OriginalURL:   "https://example2.com",
			},
		}

		err := fs.AddBatch(context.Background(), batchRequest, userID)
		require.NoError(t, err)

		// Проверяем, что все URL были добавлены
		for _, req := range batchRequest {
			result, err := fs.Get(context.Background(), req.ShortURL)
			require.NoError(t, err)
			assert.Equal(t, req.OriginalURL, result.OriginalURL)
		}
	})

	t.Run("SaveToFile and LoadFromFile", func(t *testing.T) {
		// Создаем новый FileStore для тестирования сохранения/загрузки
		testFilePath := filepath.Join(tmpDir, "test_urls.json")
		testFS := &FileStore{
			URLMapping: make(map[string]string),
			FilePath:   testFilePath,
		}

		// Добавляем тестовые данные
		testData := map[string]string{
			"test1": "https://test1.com",
			"test2": "https://test2.com",
		}
		for short, original := range testData {
			testFS.URLMapping[short] = original
		}

		// Сохраняем в файл
		err := testFS.SaveToFile()
		require.NoError(t, err)

		// Проверяем, что файл существует
		_, err = os.Stat(testFilePath)
		require.NoError(t, err)

		// Создаем новый FileStore для загрузки данных
		loadFS := &FileStore{
			URLMapping: make(map[string]string),
			FilePath:   testFilePath,
		}

		// Загружаем данные из файла
		err = loadFS.LoadFromFile()
		require.NoError(t, err)

		// Проверяем, что данные загружены корректно
		for short, original := range testData {
			loaded, ok := loadFS.URLMapping[short]
			assert.True(t, ok)
			assert.Equal(t, original, loaded)
		}
	})

	t.Run("LoadFromFile with non-existent file", func(t *testing.T) {
		nonExistentFS := &FileStore{
			URLMapping: make(map[string]string),
			FilePath:   filepath.Join(tmpDir, "nonexistent.json"),
		}

		err := nonExistentFS.LoadFromFile()
		require.NoError(t, err)
		assert.Empty(t, nonExistentFS.URLMapping)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		concurrentFS := &FileStore{
			URLMapping: make(map[string]string),
		}

		// Запускаем несколько горутин для одновременного доступа
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				shortURL := fmt.Sprintf("short%d", i)
				originalURL := fmt.Sprintf("https://example%d.com", i)
				err := concurrentFS.Add(context.Background(), shortURL, originalURL, uuid.New())
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Ждем завершения всех горутин
		for i := 0; i < 10; i++ {
			<-done
		}

		// Проверяем, что все URL были добавлены
		assert.Equal(t, 10, len(concurrentFS.URLMapping))
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		// Проверяем, что метод возвращает nil, nil
		urls, err := fs.GetUserURLs(context.Background(), userID)
		require.NoError(t, err)
		assert.Nil(t, urls)
	})

	t.Run("DeleteUserURLs", func(t *testing.T) {
		// Создаем канал с URL для удаления
		urlsToDelete := make(chan models.UserShortURL, 1)
		urlsToDelete <- models.UserShortURL{
			UserID:   userID,
			ShortURL: shortURL,
		}
		close(urlsToDelete)

		// Проверяем, что метод возвращает nil
		err := fs.DeleteUserURLs(context.Background(), urlsToDelete)
		require.NoError(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		// Проверяем, что метод возвращает ошибку
		err := fs.Ping()
		require.Error(t, err)
		assert.Equal(t, "unable to access the store", err.Error())
	})
}
