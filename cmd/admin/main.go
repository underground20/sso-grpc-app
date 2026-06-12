package main

import (
	"app/internal/admin"
	"app/internal/config"
	"app/internal/infrastructure/db"
	"app/internal/infrastructure/logging"
	"app/internal/storage"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	db, err := db.New(cfg.DatabaseUrl, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger := logging.Setup(cfg.Env)
	appCreator := storage.NewAppStorage(db)
	roleCreator := storage.NewRoleStorage(db, logger)
	adminApp := admin.New(cfg.HTTP, logger, appCreator, roleCreator)

	go func() {
		adminApp.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	adminApp.Stop()

}
