// Package app provides the main application setup and configuration for the URL shortener service.
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/learies/goShortener/internal/config"
	"github.com/learies/goShortener/internal/config/logger"
	grpcserver "github.com/learies/goShortener/internal/grpc"
	"github.com/learies/goShortener/internal/router"
	"github.com/learies/goShortener/internal/services"
	"github.com/learies/goShortener/internal/store"
	"google.golang.org/grpc/reflection"
)

// App is a struct that holds the application configuration and router.
type App struct {
	Config *config.Config
	Router *router.Router
	Server *http.Server
	// gRPC server
	GRPCServer *grpcserver.Server
}

// NewApp is a function that creates a new App instance.
func NewApp(cfg *config.Config) (*App, error) {
	router := router.NewRouter()

	store, err := store.NewStore(*cfg)
	if err != nil {
		logger.Log.Error("Failed to setup store", "error", err)
		return nil, err
	}

	urlShortener := services.NewURLShortenerService(store, cfg.BaseURL)

	if err := router.Routes(cfg, store, urlShortener); err != nil {
		logger.Log.Error("Failed to setup routes", "error", err)
		return nil, err
	}

	// Create gRPC server
	grpcServer := grpcserver.NewServer(urlShortener)
	reflection.Register(grpcServer.Server)

	return &App{
		Config:     cfg,
		Router:     router,
		GRPCServer: grpcServer,
	}, nil
}

// Run is a method that starts the server with graceful shutdown.
func (a *App) Run() error {
	logger.Log.Info("Starting server on", "address", a.Config.Address, "https", a.Config.EnableHTTPS)

	// Проверяем наличие сертификатов для HTTPS
	if a.Config.EnableHTTPS {
		if a.Config.CertFile == "" || a.Config.KeyFile == "" {
			return fmt.Errorf("certificate and key files are required for HTTPS")
		}
	}

	// Создаем HTTP сервер
	a.Server = &http.Server{
		Addr:    a.Config.Address,
		Handler: a.Router.Mux,
	}

	// Канал для сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Запускаем HTTP сервер в горутине
	go func() {
		var err error
		if a.Config.EnableHTTPS {
			err = a.Server.ListenAndServeTLS(a.Config.CertFile, a.Config.KeyFile)
		} else {
			err = a.Server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Error("HTTP server error", "error", err)
		}
	}()

	// Запускаем gRPC сервер в горутине, если включен
	if a.Config.EnableGRPC {
		logger.Log.Info("Starting gRPC server on", "address", a.Config.GRPCAddress)
		go func() {
			lis, err := net.Listen("tcp", a.Config.GRPCAddress)
			if err != nil {
				logger.Log.Error("Failed to listen", "error", err)
				return
			}
			if err := a.GRPCServer.Serve(lis); err != nil {
				logger.Log.Error("gRPC server error", "error", err)
			}
		}()
	}

	// Ждем сигнала завершения
	<-stop
	logger.Log.Info("Shutting down servers...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершаем HTTP сервер
	if err := a.Server.Shutdown(ctx); err != nil {
		logger.Log.Error("HTTP server forced to shutdown", "error", err)
		return fmt.Errorf("HTTP server forced to shutdown: %w", err)
	}

	// Завершаем gRPC сервер
	if a.Config.EnableGRPC {
		a.GRPCServer.GracefulStop()
	}

	logger.Log.Info("Servers exited properly")
	return nil
}
