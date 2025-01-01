package router

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/handler"
	internalMiddleware "github.com/learies/goShortener/internal/middleware"
	"github.com/learies/goShortener/internal/services"
)

type Router struct {
	*chi.Mux
}

func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

func (r *Router) Routes(cfg *config.Config) error {
	routes := r.Mux
	routes.Use(middleware.Recoverer)
	routes.Use(internalMiddleware.WithLogging)

	handler := handler.NewHandler()
	urlShortener := services.NewURLShortener()

	routes.Post("/", handler.CreateShortLink(cfg.BaseURL, urlShortener))

	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Error("Method not allowed", "method", r.Method, "path", r.URL.Path)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
