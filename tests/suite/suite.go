package suite

import (
	"app/internal/config"
	"context"
	"log"
	"net"
	"strconv"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5"
	sso "github.com/underground20/sso-grpc-contract/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost    = "localhost"
	databaseUrl = "postgres://admin:admin@localhost:5432/app?sslmode=disable"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient sso.AuthClient
	Connection *pgx.Conn
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := loadConfig()
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	grpcAddress := net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))

	cc, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc client connect failed: %v", err)
	}

	authClient := sso.NewAuthClient(cc)
	connection, err := pgx.Connect(ctx, databaseUrl)
	if err != nil {
		panic(err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: authClient,
		Connection: connection,
	}
}

func loadConfig() *config.Config {
	var cfg config.Config
	err := cleanenv.ReadConfig("../.env", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
