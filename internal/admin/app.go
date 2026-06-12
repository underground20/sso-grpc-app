package admin

import (
	"app/internal/admin/handler"
	"app/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type AppCreator interface {
	RegisterApp(ctx context.Context, name string, secret string) (int, error)
}

type RoleCreator interface {
	CreateRole(ctx context.Context, name string, permissions []string) (int, error)
}

type App struct {
	server *http.Server
	logger *slog.Logger
}

func New(
	cfg config.HTTPConfig,
	logger *slog.Logger,
	appCreator AppCreator,
	roleCreator RoleCreator,
) *App {
	mux := http.NewServeMux()
	mux.Handle("POST /admin/app", handler.NewAppHandler(appCreator, logger))
	mux.Handle("POST /admin/role", handler.NewRoleHandler(roleCreator, logger))

	server := &http.Server{
		Addr:         ":" + fmt.Sprintf("%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	return &App{
		server: server,
		logger: logger,
	}
}

func (a *App) Run() {
	a.logger.Info("starting admin http server", slog.String("addr", a.server.Addr))
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("admin http server failed: " + err.Error())
	}
}

func (a *App) Stop() {
	a.logger.Info("stopping admin http server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("http server shutdown failed", slog.String("error", err.Error()))
		return
	}
	a.logger.Info("http server stopped gracefully")
}
