package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type UserStorage struct {
	db     *db.Database
	logger *slog.Logger
}

func NewUserStorage(db *db.Database, logger *slog.Logger) UserStorage {
	return UserStorage{db: db, logger: logger}
}

func (s UserStorage) GetUser(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := s.db.Conn.QueryRow(
		ctx,
		`SELECT id, email, pass_hash, username, status, created_at, last_login FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID, &user.Email, &user.PassHash, &user.Username, &user.Status, &user.CreatedAt, &user.LastLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}

	rows, err := s.db.Conn.Query(ctx, `
		SELECT r.name, rp.permission
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		WHERE ur.user_id = $1
	`, user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	roles := make(map[string][]string)
	for rows.Next() {
		var roleName string
		var permission string
		if err := rows.Scan(&roleName, &permission); err != nil {
			return models.User{}, fmt.Errorf("failed to scan role/permission: %w", err)
		}

		if _, ok := roles[roleName]; !ok {
			roles[roleName] = make([]string, 0)
		}

		roles[roleName] = append(roles[roleName], permission)
	}

	if err := rows.Err(); err != nil {
		return models.User{}, fmt.Errorf("error iterating user roles: %w", err)
	}

	user.Roles = make([]string, 0, len(roles))
	user.Scopes = make([]string, 0)
	for roleName, permissions := range roles {
		user.Roles = append(user.Roles, roleName)
		user.Scopes = append(user.Scopes, permissions...)
	}

	return user, nil
}

func (s UserStorage) SaveUser(ctx context.Context, uuid uuid.UUID, email string, password []byte, username string, roles []int64) error {
	transaction, err := s.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			s.logger.Error("failed to rollback transaction", slog.String("error", rollbackErr.Error()))
		}
	}()

	var u string
	err = transaction.QueryRow(
		ctx,
		`INSERT INTO users (id, email, pass_hash, username) 
			VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING RETURNING id
		`,
		uuid,
		email,
		password,
		username,
	).Scan(&u)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserExists
		}
		return fmt.Errorf("failed to save user: %w", err)
	}

	if len(roles) > 0 {
		err = s.addRoles(ctx, transaction, uuid.String(), roles)
		if err != nil {
			return fmt.Errorf("failed to add roles: %w", err)
		}
	}

	return transaction.Commit(ctx)
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

func (s UserStorage) addRoles(ctx context.Context, transaction pgx.Tx, userID string, roles []int64) error {
	rows := make([][]interface{}, len(roles))
	for i, roleID := range roles {
		rows[i] = []interface{}{userID, roleID}
	}

	_, err := transaction.CopyFrom(
		ctx,
		pgx.Identifier{"user_roles"},
		[]string{"user_id", "role_id"},
		pgx.CopyFromRows(rows),
	)

	return err
}
