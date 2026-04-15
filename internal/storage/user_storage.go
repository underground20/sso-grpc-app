package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type UserStorage struct {
	db *db.Database
}

func NewUserStorage(db *db.Database) UserStorage {
	return UserStorage{db: db}
}

func (s UserStorage) GetUser(ctx context.Context, email string) (models.User, error) {
	rows, err := s.db.Conn.Query(
		ctx,
		`SELECT id, email, pass_hash FROM users WHERE email = $1`,
		email,
	)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to collect row: %w", err)
	}

	return user, nil
}

func (s UserStorage) SaveUser(ctx context.Context, email string, password []byte) (int64, error) {
	rows, err := s.db.Conn.Query(
		ctx,
		`INSERT INTO users (email, pass_hash) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING RETURNING id`,
		email,
		password,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	id, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserExists
		}
		return 0, fmt.Errorf("failed to collect row: %w", err)
	}

	return id, nil
}
