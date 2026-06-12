package auth

import (
	"app/internal/models"
	"app/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	jwt "github.com/underground20/sso-jwt-token/pkg/jwt/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrRolesDoesNotExists = errors.New("one or more roles does not exist")
)

type UserStorage interface {
	GetUser(ctx context.Context, email string) (models.User, error)
	SaveUser(ctx context.Context, uuid uuid.UUID, email string, password []byte, username string, roles []int64) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (models.App, error)
}

type RoleProvider interface {
	GetRoles(ctx context.Context) ([]models.Role, error)
	RolesExist(ctx context.Context, roles []int64) (bool, error)
}

type Auth struct {
	logger         *slog.Logger
	userStorage    UserStorage
	appProvider    AppProvider
	roleProvider   RoleProvider
	tokenGenerator *jwt.TokenGenerator
	passwordCost   int
}

func New(
	logger *slog.Logger,
	userStorage UserStorage,
	appProvider AppProvider,
	roleProvider RoleProvider,
	tokenGenerator *jwt.TokenGenerator,
	passwordCost int,
) *Auth {
	return &Auth{
		logger:         logger,
		userStorage:    userStorage,
		appProvider:    appProvider,
		roleProvider:   roleProvider,
		tokenGenerator: tokenGenerator,
		passwordCost:   passwordCost,
	}
}

func (a *Auth) Login(ctx context.Context, email, password string, appId int) (string, error) {
	user, err := a.userStorage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials", slog.String("error", err.Error()))

		return "", ErrInvalidCredentials
	}

	err = a.userStorage.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return "", err
	}

	app, err := a.appProvider.GetApp(ctx, appId)
	if err != nil {
		return "", err
	}

	token, err := a.tokenGenerator.Generate(
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
		return "", err
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, password, username string, roles []int64) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), a.passwordCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	newUuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %w", err)
	}

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
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	return newUuid.String(), nil
}

func (a *Auth) GetRoles(ctx context.Context) ([]models.Role, error) {
	return a.roleProvider.GetRoles(ctx)
}
