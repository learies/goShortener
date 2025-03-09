// Package app provides the main application setup and configuration for the URL shortener service.
package app

import (
	"fmt"
	"net/http"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/router"
	"github.com/learies/goShortener/internal/services"
	"github.com/learies/goShortener/internal/store"
)

// App is a struct that holds the application configuration and router.
type App struct {
	Config *config.Config
	Router *router.Router
}

// NewApp is a function that creates a new App instance.
func NewApp(cfg *config.Config) (*App, error) {
	router := router.NewRouter()

	store, err := store.NewStore(*cfg)
	if err != nil {
		logger.Log.Error("Failed to setup store", "error", err)
		return nil, err
	}

	urlShortener := services.NewURLShortener()

	if err := router.Routes(cfg, store, urlShortener); err != nil {
		logger.Log.Error("Failed to setup routes", "error", err)
		return nil, err
	}

	return &App{
		Config: cfg,
		Router: router,
	}, nil
}

// Run is a method that starts the server.
func (a *App) Run() error {
	logger.Log.Info("Starting server on", "address", a.Config.Address, "https", a.Config.EnableHTTPS)

	if a.Config.EnableHTTPS {
		if a.Config.CertFile == "" || a.Config.KeyFile == "" {
			return fmt.Errorf("certificate and key files are required for HTTPS")
		}
		return http.ListenAndServeTLS(a.Config.Address, a.Config.CertFile, a.Config.KeyFile, a.Router.Mux)
	}

	return http.ListenAndServe(a.Config.Address, a.Router.Mux)
}
