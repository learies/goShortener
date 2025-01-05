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
	"github.com/learies/goShortener/internal/store"
)

type Router struct {
	*chi.Mux
}

func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

func (r *Router) Routes(cfg *config.Config, store store.Store, urlShortener services.Shortener) error {
	routes := r.Mux
	routes.Use(middleware.Recoverer)
	routes.Use(internalMiddleware.WithLogging)
	routes.Use(internalMiddleware.GzipMiddleware)

	handler := handler.NewHandler()

	routes.Post("/", handler.CreateShortLink(store, cfg.BaseURL, urlShortener))
	routes.Get("/{shortURL}", handler.GetOriginalURL(store))
	routes.Post("/api/shorten", handler.ShortenLink(store, cfg.BaseURL, urlShortener))
	r.Get("/ping", handler.PingHandler(store))
	routes.MethodNotAllowed(methodNotAllowedHandler)

	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Error("Method not allowed", "method", r.Method, "path", r.URL.Path)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
