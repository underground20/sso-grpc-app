package auth

import (
	"app/internal/infrastructure/id"
	"app/internal/infrastructure/token"
	"app/internal/models"
	"app/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	jwt "github.com/underground20/sso-jwt-token/pkg/jwt/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrRolesDoesNotExists    = errors.New("one or more roles does not exist")
	ErrRefreshTokenIsRevoked = errors.New("refresh token is revoked")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserStorage interface {
	GetUserById(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	SaveUser(ctx context.Context, uuid uuid.UUID, email string, password []byte, username string, roles []int64) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (models.App, error)
}

type RoleProvider interface {
	GetRoles(ctx context.Context) ([]models.Role, error)
	RolesExist(ctx context.Context, roles []int64) (bool, error)
}

type RefreshTokenStorage interface {
	CreateToken(ctx context.Context, token token.RefreshToken) error
	GetToken(ctx context.Context, tokenString string) (token.RefreshToken, error)
	FetchNewToken(ctx context.Context, oldToken, newToken token.RefreshToken) error
}

type TransactionExecutor interface {
	Execute(ctx context.Context, handler func() error) error
}

type Auth struct {
	logger              *slog.Logger
	userStorage         UserStorage
	appProvider         AppProvider
	roleProvider        RoleProvider
	refreshTokenStorage RefreshTokenStorage
	transactionExecutor TransactionExecutor
	tokenGenerator      *jwt.TokenGenerator
	passwordCost        int
}

func New(
	logger *slog.Logger,
	userStorage UserStorage,
	appProvider AppProvider,
	roleProvider RoleProvider,
	refreshTokenStorage RefreshTokenStorage,
	transactionExecutor TransactionExecutor,
	tokenGenerator *jwt.TokenGenerator,
	passwordCost int,
) *Auth {
	return &Auth{
		logger:              logger,
		userStorage:         userStorage,
		appProvider:         appProvider,
		roleProvider:        roleProvider,
		refreshTokenStorage: refreshTokenStorage,
		transactionExecutor: transactionExecutor,
		tokenGenerator:      tokenGenerator,
		passwordCost:        passwordCost,
	}
}

func (a *Auth) Login(ctx context.Context, email, password string, appId int) (TokenPair, error) {
	user, err := a.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}

		return TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials", slog.String("error", err.Error()))

		return TokenPair{}, ErrInvalidCredentials
	}

	userId := id.CreateFromString(user.ID)
	refreshToken, err := token.GenerateRefreshToken(userId)
	if err != nil {
		return TokenPair{}, err
	}

	err = a.transactionExecutor.Execute(ctx, func() error {
		err = a.userStorage.UpdateLastLogin(ctx, userId)
		if err != nil {
			return err
		}

		err = a.refreshTokenStorage.CreateToken(ctx, refreshToken)
		if err != nil {
			return err
		}

		return nil
	})

	app, err := a.appProvider.GetApp(ctx, appId)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := a.tokenGenerator.Generate(
		user.ID,
		app.Name,
		jwt.Info{
			Email:  user.Email,
			Roles:  user.Roles,
			Scopes: user.Scopes,
		},
		app.Secret,
	)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken.Token}, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, password, username string, roles []int64) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), a.passwordCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	newUuid := id.Generate()
	if len(roles) > 0 {
		ok, err := a.roleProvider.RolesExist(ctx, roles)
		if err != nil {
			return "", err
		}

		if !ok {
			return "", ErrRolesDoesNotExists
		}
	}

	err = a.userStorage.SaveUser(ctx, newUuid, email, passwordHash, username, roles)
	if err != nil {
		return "", err
	}

	return newUuid.String(), nil
}

func (a *Auth) GetRoles(ctx context.Context) ([]models.Role, error) {
	return a.roleProvider.GetRoles(ctx)
}

func (a *Auth) GetNewToken(ctx context.Context, refreshTokenRaw string, appId int64) (TokenPair, error) {
	refreshToken, err := a.refreshTokenStorage.GetToken(ctx, refreshTokenRaw)
	if err != nil {
		return TokenPair{}, err
	}

	if refreshToken.Revoked {
		return TokenPair{}, ErrRefreshTokenIsRevoked
	}

	user, err := a.userStorage.GetUserById(ctx, refreshToken.UserId)
	if err != nil {
		return TokenPair{}, err
	}

	app, err := a.appProvider.GetApp(ctx, int(appId))
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := a.tokenGenerator.Generate(
		user.ID,
		app.Name,
		jwt.Info{
			Email:  user.Email,
			Roles:  user.Roles,
			Scopes: user.Scopes,
		},
		app.Secret,
	)
	if err != nil {
		return TokenPair{}, err
	}

	newRefreshTokenRaw := refreshToken.Token
	if refreshToken.ExpiresAt.Before(time.Now()) {
		newRefreshToken, err := token.GenerateRefreshToken(id.CreateFromString(user.ID))
		if err != nil {
			return TokenPair{}, err
		}

		err = a.refreshTokenStorage.FetchNewToken(ctx, refreshToken, newRefreshToken)
		if err != nil {
			return TokenPair{}, err
		}

		newRefreshTokenRaw = newRefreshToken.Token
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: newRefreshTokenRaw}, nil
}
