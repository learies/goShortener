package main

import (
	"github.com/learies/goShortener/internal/app"
	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
)

// main is the entry point for the application.
func main() {
	err := logger.NewLogger("info")
	if err != nil {
		logger.Log.Error("Error creating logger", "error", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Log.Error("Error creating config", "error", err)
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		logger.Log.Error("Error creating app", "error", err)
	}

	if err := application.Run(); err != nil {
		logger.Log.Error("Error running app", "error", err)
	}
}
