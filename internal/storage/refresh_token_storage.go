package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/infrastructure/token"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrTokenNotFound = errors.New("refresh token not found")
)

type RefreshTokenStorage struct {
	db *db.Database
}

func NewRefreshTokenStorage(db *db.Database) RefreshTokenStorage {
	return RefreshTokenStorage{db: db}
}

func (s RefreshTokenStorage) CreateToken(ctx context.Context, token token.RefreshToken) error {
	_, err := s.db.Conn.Exec(
		ctx,
		`INSERT INTO refresh_tokens(id, token, user_id, expires_at, revoked) VALUES ($1, $2, $3, $4, $5)`,
		token.Id,
		token.Token,
		token.UserId,
		token.ExpiresAt,
		token.Revoked,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", "create refresh token", err)
	}

	return nil
}

func (s RefreshTokenStorage) GetToken(ctx context.Context, tokenString string) (token.RefreshToken, error) {
	row := s.db.Conn.QueryRow(
		ctx,
		`SELECT id, token, user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1`,
		tokenString,
	)

	var refreshToken token.RefreshToken
	err := row.Scan(&refreshToken.Id, &refreshToken.Token, &refreshToken.UserId, &refreshToken.ExpiresAt, &refreshToken.Revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return refreshToken, ErrTokenNotFound
		}

		return token.RefreshToken{}, err
	}

	return refreshToken, nil
}

func (s RefreshTokenStorage) FetchNewToken(ctx context.Context, oldToken, newToken token.RefreshToken) error {
	query := `
		WITH revoked_old AS (
			UPDATE refresh_tokens
			SET revoked = true
			WHERE id = $1 AND revoked = false
		)
		INSERT INTO refresh_tokens (id, token, user_id, expires_at)
		VALUES ($2, $3, $4, $5)
	`

	_, err := s.db.Conn.Exec(
		ctx,
		query,
		oldToken.Id,
		newToken.Id,
		newToken.Token,
		newToken.UserId,
		newToken.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", "fetch new refresh token", err)
	}

	return nil
}
