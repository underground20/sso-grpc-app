package logging

import (
	"log/slog"
	"os"
)

const (
	envDev  = "dev"
	envProd = "prod"
	envTest = "test"
)

func Setup(env string) *slog.Logger {
	switch env {
	case envDev:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envTest:
		return slog.New(slog.DiscardHandler)
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
