package main

import (
	"fmt"

	"github.com/learies/goShortener/internal/app"
	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", getValueOrDefault(buildVersion))
	fmt.Printf("Build date: %s\n", getValueOrDefault(buildDate))
	fmt.Printf("Build commit: %s\n", getValueOrDefault(buildCommit))
}

func getValueOrDefault(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// main is the entry point for the application.
func main() {
	printBuildInfo()

	err := logger.NewLogger("info")
	if err != nil {
		logger.Log.Error("Error creating logger", "error", err)
		return
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Log.Error("Error creating config", "error", err)
		return
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		logger.Log.Error("Error creating app", "error", err)
		return
	}

	if err := application.Run(); err != nil {
		logger.Log.Error("Error running app", "error", err)
		return
	}
}
