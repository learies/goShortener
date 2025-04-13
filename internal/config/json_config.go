package config

import (
	"encoding/json"
	"os"
)

// JSONConfig представляет структуру конфигурации из JSON файла
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// loadJSONConfig загружает конфигурацию из JSON файла
func loadJSONConfig(configPath string) (*JSONConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	// Проверяем существование файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Открываем файл
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Декодируем JSON
	var config JSONConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeConfig объединяет значения из JSON конфигурации с существующей конфигурацией
func (c *Config) mergeConfig(jsonConfig *JSONConfig) {
	if jsonConfig == nil {
		return
	}

	if jsonConfig.ServerAddress != "" {
		c.Address = jsonConfig.ServerAddress
	}
	if jsonConfig.BaseURL != "" {
		c.BaseURL = jsonConfig.BaseURL
	}
	if jsonConfig.FileStoragePath != "" {
		c.FilePath = jsonConfig.FileStoragePath
	}
	if jsonConfig.DatabaseDSN != "" {
		c.DatabaseDSN = jsonConfig.DatabaseDSN
	}
	c.EnableHTTPS = c.EnableHTTPS || jsonConfig.EnableHTTPS
}
