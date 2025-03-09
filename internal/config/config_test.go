package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPSConfig(t *testing.T) {
	// Сохраняем оригинальные значения переменных окружения
	originalEnvVars := map[string]string{
		"ENABLE_HTTPS":   os.Getenv("ENABLE_HTTPS"),
		"CERT_FILE":      os.Getenv("CERT_FILE"),
		"KEY_FILE":       os.Getenv("KEY_FILE"),
		"SERVER_ADDRESS": os.Getenv("SERVER_ADDRESS"),
		"BASE_URL":       os.Getenv("BASE_URL"),
	}

	// Сохраняем оригинальные аргументы
	originalArgs := os.Args

	// Очищаем переменные окружения и аргументы после теста
	defer func() {
		for key, value := range originalEnvVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		args           []string
		expectedConfig Config
		wantErr        bool
	}{
		{
			name: "HTTPS enabled via flag",
			args: []string{"-s", "-cert", "/path/to/cert.pem", "-key", "/path/to/key.pem"},
			expectedConfig: Config{
				Address:     ":8080",
				BaseURL:     "https://localhost:8080",
				EnableHTTPS: true,
				CertFile:    "/path/to/cert.pem",
				KeyFile:     "/path/to/key.pem",
			},
		},
		{
			name: "HTTPS enabled via env",
			envVars: map[string]string{
				"ENABLE_HTTPS": "true",
				"CERT_FILE":    "/path/to/cert.pem",
				"KEY_FILE":     "/path/to/key.pem",
			},
			expectedConfig: Config{
				Address:     ":8080",
				BaseURL:     "https://localhost:8080",
				EnableHTTPS: true,
				CertFile:    "/path/to/cert.pem",
				KeyFile:     "/path/to/key.pem",
			},
		},
		{
			name: "HTTPS disabled (default)",
			expectedConfig: Config{
				Address:     ":8080",
				BaseURL:     "http://localhost:8080",
				EnableHTTPS: false,
			},
		},
		{
			name: "Custom base URL with HTTPS",
			args: []string{"-s", "-b", "http://mysite.com:8443", "-cert", "/cert.pem", "-key", "/key.pem"},
			expectedConfig: Config{
				Address:     ":8080",
				BaseURL:     "https://mysite.com:8443",
				EnableHTTPS: true,
				CertFile:    "/cert.pem",
				KeyFile:     "/key.pem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем переменные окружения
			for key := range originalEnvVars {
				os.Unsetenv(key)
			}

			// Устанавливаем тестовые переменные окружения
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Сбрасываем флаги перед каждым тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Устанавливаем аргументы командной строки
			os.Args = append([]string{"cmd"}, tt.args...)

			// Получаем конфигурацию
			cfg, err := NewConfig()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Проверяем основные поля конфигурации
			assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
			assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
			assert.Equal(t, tt.expectedConfig.CertFile, cfg.CertFile)
			assert.Equal(t, tt.expectedConfig.KeyFile, cfg.KeyFile)
		})
	}
}
