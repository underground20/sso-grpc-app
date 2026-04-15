package auth

import (
	"context"
	"errors"
	"log/slog"

	sso "github.com/underground20/sso-grpc-contract/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type serverAPI struct {
	sso.UnimplementedAuthServer
	auth   Auth
	logger *slog.Logger
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
}

func Register(gRPCServer *grpc.Server, auth Auth, logger *slog.Logger) {
	sso.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth, logger: logger})
}

func (s *serverAPI) Login(ctx context.Context, in *sso.LoginRequest) (*sso.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		s.logger.Error("failed to login", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &sso.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "failed to register new user")
		}

		s.logger.Error("failed to register new user", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sso.RegisterResponse{UserId: uid}, nil
}
