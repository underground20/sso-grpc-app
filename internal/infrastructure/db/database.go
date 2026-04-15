package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Conn *pgxpool.Pool
}

func New(databaseUrl string, ctx context.Context) (*Database, error) {
	conn, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		return nil, err
	}

	return &Database{Conn: conn}, nil
}
