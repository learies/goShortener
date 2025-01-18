package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/handler"
	internalMiddleware "github.com/learies/goShortener/internal/middleware"
	"github.com/learies/goShortener/internal/services"
	"github.com/learies/goShortener/internal/store"
)

// Router is a struct that wraps the chi.Mux router.
type Router struct {
	*chi.Mux
}

// NewRouter creates a new Router instance.
func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

// Routes configures the routes for the router.
func (r *Router) Routes(cfg *config.Config, store store.Store, urlShortener services.Shortener) error {
	routes := r.Mux
	routes.Use(middleware.Recoverer)
	routes.Use(internalMiddleware.WithLogging)
	routes.Use(internalMiddleware.GzipMiddleware)
	routes.Use(internalMiddleware.JWTMiddleware)

	handler := handler.NewHandler()

	routes.Post("/", handler.CreateShortLink(store, cfg.BaseURL, urlShortener))
	routes.Get("/{shortURL}", handler.GetOriginalURL(store))
	routes.Post("/api/shorten", handler.ShortenLink(store, cfg.BaseURL, urlShortener))
	routes.Get("/ping", handler.PingHandler(store))
	routes.Post("/api/shorten/batch", handler.ShortenLinkBatch(store, cfg.BaseURL, urlShortener))
	routes.Get("/api/user/urls", handler.GetUserURLs(store, cfg.BaseURL))
	routes.Delete("/api/user/urls", handler.DeleteUserURLs(store))
	routes.MethodNotAllowed(methodNotAllowedHandler)

	routes.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	routes.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	routes.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	routes.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	routes.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	routes.Handle("/debug/pprof/heap", http.HandlerFunc(pprof.Handler("heap").ServeHTTP))
	routes.Handle("/debug/pprof/goroutine", http.HandlerFunc(pprof.Handler("goroutine").ServeHTTP))
	routes.Handle("/debug/pprof/block", http.HandlerFunc(pprof.Handler("block").ServeHTTP))
	routes.Handle("/debug/pprof/mutex", http.HandlerFunc(pprof.Handler("mutex").ServeHTTP))
	routes.Handle("/debug/pprof/threadcreate", http.HandlerFunc(pprof.Handler("threadcreate").ServeHTTP))
	routes.Handle("/debug/pprof/allocs", http.HandlerFunc(pprof.Handler("allocs").ServeHTTP))
	return nil
}

// methodNotAllowedHandler is a handler that returns a 405 Method Not Allowed status.
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Error("Method not allowed", "method", r.Method, "path", r.URL.Path)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
