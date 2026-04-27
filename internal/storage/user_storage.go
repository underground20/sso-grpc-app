package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
		`SELECT id, email, pass_hash, username, status, created_at, last_login FROM users WHERE email = $1`,
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

func (s UserStorage) SaveUser(ctx context.Context, uuid uuid.UUID, email string, password []byte) error {
	rows, err := s.db.Conn.Query(
		ctx,
		`INSERT INTO users (id, email, pass_hash) 
			VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING RETURNING id
		`,
		uuid,
		email,
		password,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	_, err = pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserExists
		}
		return fmt.Errorf("failed to collect row: %w", err)
	}

	return nil
}

func (s UserStorage) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := s.db.Conn.Exec(
		ctx,
		`UPDATE users SET last_login = NOW() WHERE id = $1`,
		userID,
	)

	return err
}

func (s UserStorage) AddRole(ctx context.Context, userID string, roleID int) error {
	_, err := s.db.Conn.Exec(
		ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
		userID,
		roleID,
	)

	return err
}
