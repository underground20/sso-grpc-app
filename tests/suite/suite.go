package suite

import (
	"app/internal/infrastructure/db"
	"app/internal/infrastructure/id"
	"app/internal/infrastructure/logging"
	"app/internal/infrastructure/token"
	"app/internal/storage"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	sso "github.com/underground20/sso-grpc-contract/generated"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	AuthClient          sso.AuthClient
	db                  *db.Database
	AppStorage          storage.AppStorage
	RoleStorage         storage.RoleStorage
	UserStorage         storage.UserStorage
	RefreshTokenStorage storage.RefreshTokenStorage
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Minute*1)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	grpcPort := getEnv("GRPC_PORT", "")
	databaseUrl := getEnv("DATABASE_URL", "")
	cc, err := grpc.NewClient("app:"+grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc client connect failed: %v", err)
	}

	authClient := sso.NewAuthClient(cc)
	database, err := db.New(databaseUrl, ctx)
	if err != nil {
		panic(err)
	}

	logger := logging.Setup(getEnv("ENV", "test"))
	appStorage := storage.NewAppStorage(database)
	roleStorage := storage.NewRoleStorage(database, logger)
	userStorage := storage.NewUserStorage(database, logger)
	tokenStorage := storage.NewRefreshTokenStorage(database)

	return ctx, &Suite{
		T:                   t,
		AuthClient:          authClient,
		db:                  database,
		AppStorage:          appStorage,
		RoleStorage:         roleStorage,
		UserStorage:         userStorage,
		RefreshTokenStorage: tokenStorage,
	}
}

func (s *Suite) CreateUser(ctx context.Context, email, password, username string, roles []int64) uuid.UUID {
	userUuid, _ := uuid.NewV7()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	_ = s.UserStorage.SaveUser(ctx, userUuid, email, passwordHash, username, roles)
	return userUuid
}

func (s *Suite) CreateRefreshToken(ctx context.Context, userId uuid.UUID) token.RefreshToken {
	token, _ := token.GenerateRefreshToken(userId)
	s.RefreshTokenStorage.CreateToken(ctx, token)

	return token
}

func (s *Suite) CreateExpiredToken(ctx context.Context, userId uuid.UUID) token.RefreshToken {
	newExpiredToken := token.RefreshToken{
		Id:        id.Generate(),
		Token:     "expired",
		UserId:    userId,
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
		Revoked:   false,
	}

	s.RefreshTokenStorage.CreateToken(ctx, newExpiredToken)

	return newExpiredToken
}

func (s *Suite) CreateRevokedToken(ctx context.Context, userId uuid.UUID) token.RefreshToken {
	newExpiredToken := token.RefreshToken{
		Id:        id.Generate(),
		Token:     "test",
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Hour * 5),
		Revoked:   true,
	}

	s.RefreshTokenStorage.CreateToken(ctx, newExpiredToken)

	return newExpiredToken
}

func (s *Suite) ExistRefreshToken(ctx context.Context, userId uuid.UUID) bool {
	var exists bool
	err := s.db.Conn.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE user_id = $1)",
		userId,
	).Scan(&exists)

	if err != nil {
		return false
	}

	return exists
}

func (s *Suite) Cleanup(ctx context.Context) {
	s.T.Cleanup(func() {
		s.db.Conn.Exec(ctx, "TRUNCATE TABLE users CASCADE")
		s.db.Conn.Exec(ctx, "TRUNCATE TABLE apps CASCADE")
		s.db.Conn.Exec(ctx, "TRUNCATE TABLE roles CASCADE")
	})
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
