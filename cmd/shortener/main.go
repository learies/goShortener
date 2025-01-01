package main

import (
	"log"

	"github.com/learies/goShortener/internal/app"
)

func main() {
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
