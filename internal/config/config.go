// Package config provides configuration management functionality for the URL shortener service.
package config

import (
	"flag"
	"os"
)

// Config is a struct that holds the configuration for the application.
type Config struct {
	Address     string
	BaseURL     string
	FilePath    string
	DatabaseDSN string
}

// getEnv is a function that retrieves the value of an environment variable.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// NewConfig is a function that creates a new Config instance.
func NewConfig() (*Config, error) {
	defaultAddress := ":8080"
	defaultBaseURL := "http://localhost" + defaultAddress
	var defaultFilePath string
	var defaultDatabaseDSN string

	envAddress := getEnv("SERVER_ADDRESS", defaultAddress)
	envBaseURL := getEnv("BASE_URL", defaultBaseURL)
	envFilePath := getEnv("FILE_STORAGE_PATH", defaultFilePath)
	envDatabaseDSN := getEnv("DATABASE_DSN", defaultDatabaseDSN)

	address := flag.String("a", envAddress, "address to start the HTTP server")
	baseURL := flag.String("b", envBaseURL, "base URL for shortened URLs")
	filePath := flag.String("f", envFilePath, "path to the file for storing URL data")
	databaseDSN := flag.String("d", envDatabaseDSN, "database DSN")

	flag.Parse()

	return &Config{
		Address:     *address,
		BaseURL:     *baseURL,
		FilePath:    *filePath,
		DatabaseDSN: *databaseDSN,
	}, nil
}
