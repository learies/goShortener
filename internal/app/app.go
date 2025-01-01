package app

import (
	"log"
	"net/http"

	"github.com/learies/goShortener/internal/config"
)

type App struct {
	Config *config.Config
}

func NewApp(cfg *config.Config) (*App, error) {
	return &App{
		Config: cfg,
	}, nil
}

func (a *App) Run() error {
	log.Println("Starting server on:", a.Config.Address)
	return http.ListenAndServe(a.Config.Address, nil)
}
