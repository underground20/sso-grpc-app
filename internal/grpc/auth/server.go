package auth

import (
	"app/internal/auth"
	"app/internal/models"
	"app/internal/storage"
	"context"
	"errors"
	"log/slog"

	sso "github.com/underground20/sso-grpc-contract/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	sso.UnimplementedAuthServer
	auth         Auth
	roleProvider RoleProvider
	logger       *slog.Logger
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
	) (userID string, err error)
}

type RoleProvider interface {
	GetRoles(ctx context.Context) ([]models.Role, error)
}

func Register(gRPCServer *grpc.Server, auth Auth, roleProvider RoleProvider, logger *slog.Logger) {
	sso.RegisterAuthServer(gRPCServer, newServerAPI(auth, roleProvider, logger))
}

func (s *serverAPI) Login(ctx context.Context, in *sso.LoginRequest) (*sso.LoginResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		s.logger.Error("failed to login", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &sso.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already registered")
		}

		s.logger.Error("failed to register new user", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sso.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) GetRoles(ctx context.Context, _ *sso.GetRolesRequest) (*sso.GetRolesResponse, error) {
	roles, err := s.roleProvider.GetRoles(ctx)
	if err != nil {
		s.logger.Error("failed to get roles", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "failed to get roles")
	}

	rolesList := make([]*sso.Role, 0, len(roles))
	for _, role := range roles {
		rolesList = append(rolesList, &sso.Role{Name: role.Name, Permissions: role.Permissions})
	}

	return &sso.GetRolesResponse{Roles: rolesList}, nil
}

func newServerAPI(auth Auth, roleProvider RoleProvider, logger *slog.Logger) *serverAPI {
	return &serverAPI{
		auth:         auth,
		roleProvider: roleProvider,
		logger:       logger,
	}
}
