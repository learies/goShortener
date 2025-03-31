// Package config provides configuration management functionality for the URL shortener service.
package config

import (
	"flag"
	"os"
	"strings"
)

// Config is a struct that holds the configuration for the application.
type Config struct {
	Address       string
	BaseURL       string
	FilePath      string
	DatabaseDSN   string
	EnableHTTPS   bool
	CertFile      string
	KeyFile       string
	TrustedSubnet string
	// gRPC server configuration
	GRPCAddress string
	EnableGRPC  bool
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
	defaultGRPCAddress := ":50051"
	var defaultFilePath string
	var defaultDatabaseDSN string
	var defaultCertFile string
	var defaultKeyFile string
	var defaultTrustedSubnet string

	// Определяем все флаги
	configPath := flag.String("c", getEnv("CONFIG", ""), "path to configuration file")
	address := flag.String("a", "", "address to start the HTTP server")
	baseURL := flag.String("b", "", "base URL for shortened URLs")
	filePath := flag.String("f", "", "path to the file for storing URL data")
	databaseDSN := flag.String("d", "", "database DSN")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS server")
	certFile := flag.String("cert", "", "path to SSL certificate file")
	keyFile := flag.String("key", "", "path to SSL private key file")
	trustedSubnet := flag.String("t", "", "trusted subnet in CIDR format")
	// gRPC flags
	grpcAddress := flag.String("grpc-addr", "", "address to start the gRPC server")
	enableGRPC := flag.Bool("grpc", false, "enable gRPC server")

	// Парсим флаги
	flag.Parse()

	// Загружаем конфигурацию из JSON файла
	jsonConfig, err := loadJSONConfig(*configPath)
	if err != nil {
		return nil, err
	}

	// Создаем базовую конфигурацию с дефолтными значениями
	cfg := &Config{
		Address:       defaultAddress,
		BaseURL:       defaultBaseURL,
		FilePath:      defaultFilePath,
		DatabaseDSN:   defaultDatabaseDSN,
		EnableHTTPS:   false,
		CertFile:      defaultCertFile,
		KeyFile:       defaultKeyFile,
		TrustedSubnet: defaultTrustedSubnet,
		GRPCAddress:   defaultGRPCAddress,
		EnableGRPC:    false,
	}

	// Применяем значения из JSON конфигурации (низший приоритет)
	cfg.mergeConfig(jsonConfig)

	// Применяем значения из переменных окружения (средний приоритет)
	if envAddress := getEnv("SERVER_ADDRESS", ""); envAddress != "" {
		cfg.Address = envAddress
	}
	if envBaseURL := getEnv("BASE_URL", ""); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envFilePath := getEnv("FILE_STORAGE_PATH", ""); envFilePath != "" {
		cfg.FilePath = envFilePath
	}
	if envDatabaseDSN := getEnv("DATABASE_DSN", ""); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}
	if envEnableHTTPS := getEnv("ENABLE_HTTPS", ""); envEnableHTTPS == "true" {
		cfg.EnableHTTPS = true
	}
	if envCertFile := getEnv("CERT_FILE", ""); envCertFile != "" {
		cfg.CertFile = envCertFile
	}
	if envKeyFile := getEnv("KEY_FILE", ""); envKeyFile != "" {
		cfg.KeyFile = envKeyFile
	}
	if envTrustedSubnet := getEnv("TRUSTED_SUBNET", ""); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}
	// gRPC environment variables
	if envGRPCAddress := getEnv("GRPC_SERVER_ADDRESS", ""); envGRPCAddress != "" {
		cfg.GRPCAddress = envGRPCAddress
	}
	if envEnableGRPC := getEnv("ENABLE_GRPC", ""); envEnableGRPC == "true" {
		cfg.EnableGRPC = true
	}

	// Применяем значения из флагов (высший приоритет)
	if *address != "" {
		cfg.Address = *address
	}
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *filePath != "" {
		cfg.FilePath = *filePath
	}
	if *databaseDSN != "" {
		cfg.DatabaseDSN = *databaseDSN
	}
	if *enableHTTPS {
		cfg.EnableHTTPS = true
	}
	if *certFile != "" {
		cfg.CertFile = *certFile
	}
	if *keyFile != "" {
		cfg.KeyFile = *keyFile
	}
	if *trustedSubnet != "" {
		cfg.TrustedSubnet = *trustedSubnet
	}
	// gRPC flags
	if *grpcAddress != "" {
		cfg.GRPCAddress = *grpcAddress
	}
	if *enableGRPC {
		cfg.EnableGRPC = true
	}

	// Update baseURL scheme if HTTPS is enabled
	if cfg.EnableHTTPS && strings.HasPrefix(cfg.BaseURL, "http://") {
		cfg.BaseURL = "https://" + strings.TrimPrefix(cfg.BaseURL, "http://")
	}

	return cfg, nil
}
