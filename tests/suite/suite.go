package suite

import (
	"app/internal/infrastructure/db"
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
	AuthClient  sso.AuthClient
	db          *db.Database
	AppStorage  storage.AppStorage
	RoleStorage storage.RoleStorage
	UserStorage storage.UserStorage
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Minute*1)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	grpcAddress := getEnv("LOCAL_GRPC_ADDRESS", "localhost:44044")
	databaseUrl := getEnv("LOCAL_DATABASE_URL", "postgres://admin:admin@localhost:5432/app_test?sslmode=disable")
	cc, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc client connect failed: %v", err)
	}

	authClient := sso.NewAuthClient(cc)
	database, err := db.New(databaseUrl, ctx)
	if err != nil {
		panic(err)
	}

	appStorage := storage.NewAppStorage(database)
	roleStorage := storage.NewRoleStorage(database)
	userStorage := storage.NewUserStorage(database)

	return ctx, &Suite{
		T:           t,
		AuthClient:  authClient,
		db:          database,
		AppStorage:  appStorage,
		RoleStorage: roleStorage,
		UserStorage: userStorage,
	}
}

func (s *Suite) CreateUser(ctx context.Context, email, password string) string {
	userUuid, _ := uuid.NewV7()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	_ = s.UserStorage.SaveUser(ctx, userUuid, email, passwordHash)
	return userUuid.String()
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
