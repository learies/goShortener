package config

type Config struct {
	Address string
}

func NewConfig() (*Config, error) {
	defaultAddress := "localhost:8080"

	return &Config{
		Address: defaultAddress,
	}, nil
}
