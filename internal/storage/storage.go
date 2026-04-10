package storage

import (
	"app/internal/domain/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

type Storage struct {
	db *Database
}

func NewStorage(db *Database) Storage {
	return Storage{db: db}
}

func (s Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	rows, err := s.db.conn.Query(
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

func (s Storage) SaveUser(ctx context.Context, email string, password []byte) (int64, error) {
	rows, err := s.db.conn.Query(
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

func (s Storage) GetApp(ctx context.Context, appID int) (models.App, error) {
	rows, err := s.db.conn.Query(
		ctx,
		`SELECT id, name, secret FROM apps WHERE id = $1`,
		appID,
	)

	if err != nil {
		return models.App{}, fmt.Errorf("failed to query app: %w", err)
	}
	defer rows.Close()

	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("failed to collect row: %w", err)
	}

	return app, nil
}
