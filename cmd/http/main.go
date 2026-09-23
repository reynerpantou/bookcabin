package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/reynerpantou/bookcabin/common/config"
	"github.com/reynerpantou/bookcabin/internal/delivery"
	"github.com/reynerpantou/bookcabin/internal/repository"
	"github.com/reynerpantou/bookcabin/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// prepare application context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	// initialize repositories
	repositories, err := repository.NewRepositories(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}
	// initialize usecases
	useCases, err := usecase.NewUseCases(ctx, cfg, repositories)
	if err != nil {
		return fmt.Errorf("failed to initialize usecases: %w", err)
	}
	// initialize delivery
	deliveryHTTP := delivery.NewHTTP(useCases)
	// initialize http server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      deliveryHTTP.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	// start server in background
	serverErr := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()
	// wait for shutdown signal or server failure
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		return fmt.Errorf("serve HTTP: %w", err)
	}
	// graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}
