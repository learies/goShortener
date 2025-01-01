package main

import (
	"log"

	"github.com/learies/goShortener/internal/app"
	"github.com/learies/goShortener/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
