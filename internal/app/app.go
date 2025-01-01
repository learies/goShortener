package app

import (
	"net/http"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
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
	logger.Log.Info("Starting server on", "address", a.Config.Address)
	return http.ListenAndServe(a.Config.Address, nil)
}
