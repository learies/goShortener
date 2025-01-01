package config

type Config struct {
	Address string
	BaseURL string
}

func NewConfig() (*Config, error) {
	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"

	return &Config{
		Address: defaultAddress,
		BaseURL: defaultBaseURL,
	}, nil
}
