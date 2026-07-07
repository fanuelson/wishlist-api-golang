package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fanuelson/wishlist-api/internal/config"
	"github.com/fanuelson/wishlist-api/internal/wishlist/adapter"
	"github.com/fanuelson/wishlist-api/internal/wishlist/application"
	"github.com/fanuelson/wishlist-api/internal/wishlist/domain"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	if err := run(logger); err != nil {
		logger.Error("application terminated", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	postgresAdapter := adapter.NewPostgresAdapter(pool)
	policy := domain.NewLimitPolicy(cfg.MaxItems)

	addUseCase := application.NewAddProductToWishlistUseCase(postgresAdapter, postgresAdapter, policy)

	handler := adapter.NewHandler(addUseCase, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := handlers.LoggingHandler(os.Stdout, mux)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: loggedHandler}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr, "maxItems", cfg.MaxItems)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	logger.Info("server stopped gracefully")
	return nil
}
