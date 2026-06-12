package auth

import (
	"app/internal/auth"
	"app/internal/models"
	"app/internal/storage"
	"context"
	"errors"
	"log/slog"

	sso "github.com/underground20/sso-grpc-contract/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	sso.UnimplementedAuthServer
	auth   Auth
	logger *slog.Logger
}

func NewServer(auth Auth, logger *slog.Logger) *Server {
	return &Server{
		auth:   auth,
		logger: logger,
	}
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
		username string,
		roles []int64,
	) (userID string, err error)
	GetRoles(ctx context.Context) ([]models.Role, error)
}

func (s *Server) Login(ctx context.Context, in *sso.LoginRequest) (*sso.LoginResponse, error) {
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
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		s.logger.Error("failed to login", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &sso.LoginResponse{Token: token}, nil
}

func (s *Server) Register(ctx context.Context, in *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword(), in.GetUsername(), in.GetRoles())
	if err != nil {
		if errors.Is(err, auth.ErrRolesDoesNotExists) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already registered")
		}

		s.logger.Error("failed to register new user", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "failed to register new user")
	}

	return &sso.RegisterResponse{UserId: uid}, nil
}

func (s *Server) GetRoles(ctx context.Context, _ *sso.GetRolesRequest) (*sso.GetRolesResponse, error) {
	roles, err := s.auth.GetRoles(ctx)
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
