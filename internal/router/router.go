package router

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/handler"
	internalMiddleware "github.com/learies/goShortener/internal/middleware"
)

type Router struct {
	*chi.Mux
}

func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

func (r *Router) Routes() error {
	routes := r.Mux
	routes.Use(middleware.Recoverer)
	routes.Use(internalMiddleware.WithLogging)

	handler := handler.NewHandler()

	routes.Post("/", handler.CreateShortLink())

	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Error("Method not allowed", "method", r.Method, "path", r.URL.Path)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
