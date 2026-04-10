package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	conn *pgxpool.Pool
}

func New(databaseUrl string, ctx context.Context) (*Database, error) {
	conn, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		return nil, err
	}

	return &Database{conn: conn}, nil
}
