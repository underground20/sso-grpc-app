package auth

import (
	"app/internal/auth"
	"app/internal/models"
	"context"
	"log/slog"

	sso "github.com/underground20/sso-grpc-contract/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	sso.UnimplementedAuthServer
	auth      Auth
	validator Validator
	logger    *slog.Logger
}

func NewServer(auth Auth, logger *slog.Logger, validator Validator) *Server {
	return &Server{
		auth:      auth,
		logger:    logger,
		validator: validator,
	}
}

type Auth interface {
	Login(ctx context.Context, command auth.LoginCommand) (token auth.TokenPair, err error)
	RegisterNewUser(ctx context.Context, command auth.RegisterCommand) (userID string, err error)
	GetRoles(ctx context.Context) ([]models.Role, error)
	GetNewToken(ctx context.Context, refreshToken string, appId int64) (auth.TokenPair, error)
}

func (s *Server) Login(ctx context.Context, in *sso.LoginRequest) (*sso.LoginResponse, error) {
	command := auth.LoginCommand{
		Email:    in.GetEmail(),
		Password: in.GetPassword(),
		AppID:    int(in.GetAppId()),
	}

	err := s.validator.Validate(command)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.auth.Login(ctx, command)
	if err != nil {
		gRPCErr, isInternal := ToGRPCErr(err, "failed to login")
		if isInternal {
			s.logger.Error("failed to login", slog.String("error", err.Error()))
		}

		return nil, gRPCErr
	}

	return &sso.LoginResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (s *Server) Register(ctx context.Context, in *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	command := auth.RegisterCommand{
		Email:    in.GetEmail(),
		Password: in.GetPassword(),
		Username: in.GetUsername(),
		Roles:    in.GetRoles(),
	}

	err := s.validator.Validate(command)
	if err != nil {
		return nil, err
	}

	uid, err := s.auth.RegisterNewUser(ctx, command)
	if err != nil {
		gRPCErr, isInternal := ToGRPCErr(err, "failed to register new user")
		if isInternal {
			s.logger.Error("failed to register new user", slog.String("error", err.Error()))
		}

		return nil, gRPCErr

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

func (s *Server) GetNewToken(ctx context.Context, in *sso.GetNewTokenRequest) (*sso.GetNewTokenResponse, error) {
	if in.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty refresh token")
	}

	tokenPair, err := s.auth.GetNewToken(ctx, in.GetRefreshToken(), in.GetAppId())
	if err != nil {
		gRPCErr, isInternal := ToGRPCErr(err, "failed to get new token")
		if isInternal {
			s.logger.Error("failed to get new user", slog.String("error", err.Error()))
		}

		return nil, gRPCErr
	}

	return &sso.GetNewTokenResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}
