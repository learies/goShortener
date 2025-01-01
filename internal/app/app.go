package app

import (
	"log"
	"net/http"
)

type App struct{}

func NewApp() (*App, error) {
	return &App{}, nil
}

func (a *App) Run() error {
	log.Println("Starting server on :8080")
	return http.ListenAndServe(":8080", nil)
}
