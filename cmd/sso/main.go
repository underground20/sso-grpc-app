package main

import (
	"app/internal/app"
	"app/internal/config"
	"app/internal/services/auth"
	"app/internal/storage"
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	db, err := storage.New(cfg.DatabaseUrl, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger := setupLogger(cfg.Env)
	store := storage.NewStorage(db)
	auth := auth.New(logger, store, store, cfg.TokenTTL)
	app := app.New(logger, auth, cfg.GRPC.Port, cfg.TokenTTL)

	go func() {
		app.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	app.Stop()
	logger.Info("Graceful shutdown")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envDev:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
