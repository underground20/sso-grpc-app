package main

import (
	"app/internal/app"
	"app/internal/auth"
	"app/internal/config"
	"app/internal/infrastructure/db"
	"app/internal/infrastructure/logging"
	"app/internal/storage"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/underground20/sso-jwt-token/pkg/jwt/user"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	db, err := db.New(cfg.DatabaseUrl, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger := logging.Setup(cfg.Env)
	userStorage := storage.NewUserStorage(db, logger)
	appStorage := storage.NewAppStorage(db)
	tokenGenerator, err := user.NewTokenGenerator(cfg.TokenTTL)
	roleProvider := storage.NewRoleStorage(db, logger)
	if err != nil {
		log.Fatalf("Failed to create token generator: %v", err)
	}

	auth := auth.New(logger, userStorage, appStorage, roleProvider, tokenGenerator, cfg.PasswordCost)
	grpcApp := app.New(logger, auth, cfg.GRPC.Port, cfg.TokenTTL)

	go func() {
		grpcApp.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	grpcApp.Stop()
	logger.Info("Graceful shutdown")
}
