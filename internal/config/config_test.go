package config

import (
	"flag"
	"os"
	"path/filepath"
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

func TestJSONConfig(t *testing.T) {
	// Создаем временную директорию для тестового файла конфигурации
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Создаем тестовый файл конфигурации
	configContent := `{
		"server_address": "localhost:9090",
		"base_url": "http://localhost:9090",
		"file_storage_path": "/path/to/file.db",
		"database_dsn": "postgres://user:pass@localhost:5432/db",
		"enable_https": true
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Сохраняем оригинальные значения
	originalEnvVars := map[string]string{
		"CONFIG": os.Getenv("CONFIG"),
	}
	originalArgs := os.Args

	// Очищаем после теста
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
		configPath     string
		envVars        map[string]string
		args           []string
		expectedConfig Config
		wantErr        bool
	}{
		{
			name:       "Load config from file via flag",
			configPath: configFile,
			args:       []string{"-c", configFile},
			expectedConfig: Config{
				Address:     "localhost:9090",
				BaseURL:     "https://localhost:9090",
				FilePath:    "/path/to/file.db",
				DatabaseDSN: "postgres://user:pass@localhost:5432/db",
				EnableHTTPS: true,
			},
		},
		{
			name:       "Load config from file via env",
			configPath: configFile,
			envVars: map[string]string{
				"CONFIG": configFile,
			},
			expectedConfig: Config{
				Address:     "localhost:9090",
				BaseURL:     "https://localhost:9090",
				FilePath:    "/path/to/file.db",
				DatabaseDSN: "postgres://user:pass@localhost:5432/db",
				EnableHTTPS: true,
			},
		},
		{
			name:       "Config file not found",
			configPath: "/nonexistent/config.json",
			args:       []string{"-c", "/nonexistent/config.json"},
			expectedConfig: Config{
				Address: ":8080",
				BaseURL: "http://localhost:8080",
			},
		},
		{
			name:       "Invalid JSON config",
			configPath: filepath.Join(tmpDir, "invalid.json"),
			args:       []string{"-c", filepath.Join(tmpDir, "invalid.json")},
			wantErr:    true,
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

			// Сбрасываем флаги
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Устанавливаем аргументы командной строки
			os.Args = append([]string{"cmd"}, tt.args...)

			// Для теста с невалидным JSON создаем файл
			if tt.name == "Invalid JSON config" {
				err := os.WriteFile(tt.configPath, []byte("{invalid json}"), 0644)
				require.NoError(t, err)
			}

			// Получаем конфигурацию
			cfg, err := NewConfig()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Проверяем поля конфигурации
			assert.Equal(t, tt.expectedConfig.Address, cfg.Address)
			assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
			assert.Equal(t, tt.expectedConfig.FilePath, cfg.FilePath)
			assert.Equal(t, tt.expectedConfig.DatabaseDSN, cfg.DatabaseDSN)
			assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
		})
	}
}

func TestConfigPriority(t *testing.T) {
	// Создаем временную директорию для тестового файла конфигурации
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Создаем тестовый файл конфигурации
	configContent := `{
		"server_address": "localhost:9090",
		"base_url": "http://localhost:9090",
		"file_storage_path": "/path/to/file.db",
		"database_dsn": "postgres://user:pass@localhost:5432/db",
		"enable_https": true
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Сохраняем оригинальные значения
	originalEnvVars := map[string]string{
		"CONFIG":            os.Getenv("CONFIG"),
		"SERVER_ADDRESS":    os.Getenv("SERVER_ADDRESS"),
		"BASE_URL":          os.Getenv("BASE_URL"),
		"FILE_STORAGE_PATH": os.Getenv("FILE_STORAGE_PATH"),
		"DATABASE_DSN":      os.Getenv("DATABASE_DSN"),
		"ENABLE_HTTPS":      os.Getenv("ENABLE_HTTPS"),
	}
	originalArgs := os.Args

	// Очищаем после теста
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
		configPath     string
		envVars        map[string]string
		args           []string
		expectedConfig Config
	}{
		{
			name:       "Flags override JSON config",
			configPath: configFile,
			args: []string{
				"-c", configFile,
				"-a", "localhost:8080",
				"-b", "http://localhost:8080",
				"-f", "/custom/path.db",
				"-d", "postgres://custom:pass@localhost:5432/db",
				"-s",
			},
			expectedConfig: Config{
				Address:     "localhost:8080",
				BaseURL:     "https://localhost:8080",
				FilePath:    "/custom/path.db",
				DatabaseDSN: "postgres://custom:pass@localhost:5432/db",
				EnableHTTPS: true,
			},
		},
		{
			name:       "Env vars override JSON config",
			configPath: configFile,
			envVars: map[string]string{
				"CONFIG":            configFile,
				"SERVER_ADDRESS":    "localhost:8080",
				"BASE_URL":          "http://localhost:8080",
				"FILE_STORAGE_PATH": "/custom/path.db",
				"DATABASE_DSN":      "postgres://custom:pass@localhost:5432/db",
				"ENABLE_HTTPS":      "true",
			},
			expectedConfig: Config{
				Address:     "localhost:8080",
				BaseURL:     "https://localhost:8080",
				FilePath:    "/custom/path.db",
				DatabaseDSN: "postgres://custom:pass@localhost:5432/db",
				EnableHTTPS: true,
			},
		},
		{
			name:       "Flags override env vars",
			configPath: configFile,
			envVars: map[string]string{
				"CONFIG":            configFile,
				"SERVER_ADDRESS":    "localhost:8080",
				"BASE_URL":          "http://localhost:8080",
				"FILE_STORAGE_PATH": "/custom/path.db",
				"DATABASE_DSN":      "postgres://custom:pass@localhost:5432/db",
				"ENABLE_HTTPS":      "true",
			},
			args: []string{
				"-c", configFile,
				"-a", "localhost:9090",
				"-b", "http://localhost:9090",
				"-f", "/other/path.db",
				"-d", "postgres://other:pass@localhost:5432/db",
				"-s",
			},
			expectedConfig: Config{
				Address:     "localhost:9090",
				BaseURL:     "https://localhost:9090",
				FilePath:    "/other/path.db",
				DatabaseDSN: "postgres://other:pass@localhost:5432/db",
				EnableHTTPS: true,
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

			// Сбрасываем флаги
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Устанавливаем аргументы командной строки
			os.Args = append([]string{"cmd"}, tt.args...)

			// Получаем конфигурацию
			cfg, err := NewConfig()
			require.NoError(t, err)

			// Проверяем поля конфигурации
			assert.Equal(t, tt.expectedConfig.Address, cfg.Address)
			assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
			assert.Equal(t, tt.expectedConfig.FilePath, cfg.FilePath)
			assert.Equal(t, tt.expectedConfig.DatabaseDSN, cfg.DatabaseDSN)
			assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
		})
	}
}
