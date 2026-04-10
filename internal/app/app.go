package app

import (
	"app/internal/grpc/auth"
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	logger     *slog.Logger
	gRPCServer *grpc.Server
	port       int
	tokenTTL   time.Duration
}

func New(logger *slog.Logger, authServer auth.Auth, port int, tokenTTL time.Duration) *App {
	recoveryOptions := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) error {
			logger.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	loggingOptions := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOptions...),
		logging.UnaryServerInterceptor(interceptorLogger(logger), loggingOptions...),
	))

	auth.Register(gRPCServer, authServer, logger)

	return &App{
		logger:     logger,
		gRPCServer: gRPCServer,
		port:       port,
		tokenTTL:   tokenTTL,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}

func (a *App) run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	a.logger.Info("gRPC server started on port", slog.String("port", listener.Addr().String()))

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
