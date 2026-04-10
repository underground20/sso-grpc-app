package auth

import (
	"app/internal/domain/models"
	"app/internal/infrastructure/jwt"
	"app/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStorage interface {
	GetUser(ctx context.Context, email string) (models.User, error)
	SaveUser(ctx context.Context, email string, password []byte) (int64, error)
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (models.App, error)
}

type Auth struct {
	logger      *slog.Logger
	userStorage UserStorage
	appProvider AppProvider
	tokenTTL    time.Duration
}

func New(
	logger *slog.Logger,
	userStorage UserStorage,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		logger:      logger,
		userStorage: userStorage,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email, password string, appId int) (string, error) {
	user, err := a.userStorage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials", slog.String("error", err.Error()))

		return "", ErrInvalidCredentials
	}

	app, err := a.appProvider.GetApp(ctx, appId)
	if err != nil {
		return "", err
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to generate password hash: %w", err)
	}

	id, err := a.userStorage.SaveUser(ctx, email, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}
