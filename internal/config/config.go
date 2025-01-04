package config

import (
	"flag"
	"os"
)

type Config struct {
	Address  string
	BaseURL  string
	FilePath string
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func NewConfig() (*Config, error) {
	defaultAddress := ":8080"
	defaultBaseURL := "http://localhost" + defaultAddress
	var defaultFilePath string

	envAddress := getEnv("SERVER_ADDRESS", defaultAddress)
	envBaseURL := getEnv("BASE_URL", defaultBaseURL)
	envFilePath := getEnv("FILE_STORAGE_PATH", defaultFilePath)

	address := flag.String("a", envAddress, "address to start the HTTP server")
	baseURL := flag.String("b", envBaseURL, "base URL for shortened URLs")
	filePath := flag.String("f", envFilePath, "path to the file for storing URL data")

	flag.Parse()

	return &Config{
		Address:  *address,
		BaseURL:  *baseURL,
		FilePath: *filePath,
	}, nil
}
