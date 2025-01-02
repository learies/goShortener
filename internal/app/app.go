package app

import (
	"net/http"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/router"
	"github.com/learies/goShortener/internal/services"
)

type App struct {
	Config *config.Config
	Router *router.Router
}

func NewApp(cfg *config.Config) (*App, error) {
	router := router.NewRouter()

	urlShortener := services.NewURLShortener()

	if err := router.Routes(cfg, urlShortener); err != nil {
		logger.Log.Error("Failed to setup routes", "error", err)
		return nil, err
	}

	return &App{
		Config: cfg,
		Router: router,
	}, nil
}

func (a *App) Run() error {
	logger.Log.Info("Starting server on", "address", a.Config.Address)
	return http.ListenAndServe(a.Config.Address, a.Router.Mux)
}
