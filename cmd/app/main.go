package main

import (
	"context"
	_ "context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ReilEgor/RepoNotifier/internal/config"
	"github.com/caarlos0/env/v11"
)

// Swagger Metadata for API Documentation
// @title
// @version         1.0
// @description
// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		logger.Error("failed to load config",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app, cleanup, err := InitializeApp(
		ctx, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, 0, cfg.DSN,
	)
	defer cleanup()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Server.Run(ctx, ":"+string(cfg.HTTPPort))
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down gracefully")
		if err := <-errCh; err != nil {
			logger.Error("server shutdown error", slog.Any("error", err))
		}
		logger.Info("server stopped")
	case err := <-errCh:
		logger.Error("server stopped unexpectedly", slog.Any("error", err))
	}
}
