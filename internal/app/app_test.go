package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store"
)

// MockStore реализует интерфейс store.Store для тестирования
type MockStore struct{}

func (m *MockStore) Add(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
	return nil
}

func (m *MockStore) Get(ctx context.Context, shortURL string) (models.ShortenStore, error) {
	return models.ShortenStore{}, nil
}

func (m *MockStore) AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error {
	return nil
}

func (m *MockStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	return nil, nil
}

func (m *MockStore) DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
	return nil
}

func (m *MockStore) Ping() error {
	return nil
}

func (m *MockStore) GetStats(ctx context.Context) (int, int, error) {
	return 0, 0, nil
}

func init() {
	err := logger.NewLogger("info")
	if err != nil {
		panic(err)
	}
}

// generateTestCert создает самоподписанный сертификат для тестирования
func generateTestCert(certFile, keyFile string) error {
	// Генерируем команду для создания самоподписанного сертификата
	cmd := fmt.Sprintf("openssl req -x509 -newkey rsa:4096 -keyout %s -out %s -days 1 -nodes "+
		"-subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'",
		keyFile, certFile)

	// Запускаем команду
	err := exec.Command("sh", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("failed to generate test certificate: %w", err)
	}

	return nil
}

func TestAppRun(t *testing.T) {
	// Создаем временные файлы для сертификатов
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")

	// Генерируем тестовые сертификаты
	err := generateTestCert(certFile, keyFile)
	require.NoError(t, err)

	// Создаем мок хранилища
	mockStore := &MockStore{}

	tests := []struct {
		name        string
		config      *config.Config
		wantErr     bool
		checkServer bool
	}{
		{
			name: "HTTP server",
			config: &config.Config{
				Address: "localhost:8081",
				BaseURL: "http://localhost:8081",
			},
			checkServer: true,
		},
		{
			name: "HTTPS server",
			config: &config.Config{
				Address:     "localhost:8082",
				BaseURL:     "https://localhost:8082",
				EnableHTTPS: true,
				CertFile:    certFile,
				KeyFile:     keyFile,
			},
			checkServer: true,
		},
		{
			name: "HTTPS without cert files",
			config: &config.Config{
				Address:     "localhost:8083",
				BaseURL:     "https://localhost:8083",
				EnableHTTPS: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подменяем конструктор хранилища
			originalConstructor := store.NewStore
			store.NewStore = func(cfg config.Config) (store.Store, error) {
				return mockStore, nil
			}
			t.Cleanup(func() {
				store.NewStore = originalConstructor
			})

			app, err := NewApp(tt.config)
			require.NoError(t, err)

			// Запускаем сервер в горутине
			errCh := make(chan error)
			go func() {
				errCh <- app.Run()
			}()

			if tt.checkServer {
				// Даем серверу время на запуск
				time.Sleep(100 * time.Millisecond)

				// Проверяем доступность сервера
				client := &http.Client{
					Timeout: 1 * time.Second,
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true, // Для тестового самоподписанного сертификата
						},
					},
				}

				resp, err := client.Get(tt.config.BaseURL + "/ping")
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// Для HTTPS проверяем, что используется TLS
				if tt.config.EnableHTTPS {
					require.NotNil(t, resp.TLS)
					require.NotEmpty(t, resp.TLS.PeerCertificates)
					cert := resp.TLS.PeerCertificates[0]
					assert.Contains(t, cert.DNSNames, "localhost")
				}
			} else if tt.wantErr {
				// Ждем ошибку от сервера
				select {
				case err := <-errCh:
					assert.Error(t, err)
				case <-time.After(100 * time.Millisecond):
					t.Error("Expected error but got none")
				}
			}
		})
	}
}
