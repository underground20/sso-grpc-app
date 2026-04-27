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
	ErrAppNotFound = errors.New("app not found")
)

type AppStorage struct {
	db *db.Database
}

func NewAppStorage(db *db.Database) AppStorage {
	return AppStorage{db: db}
}

func (s AppStorage) GetApp(ctx context.Context, appID int) (models.App, error) {
	rows, err := s.db.Conn.Query(
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

func (s AppStorage) RegisterApp(ctx context.Context, name, secret string) (int, error) {
	const op = "storage.app.RegisterApp"
	var id int
	query := `INSERT INTO apps(name, secret) VALUES ($1, $2) RETURNING id`
	err := s.db.Conn.QueryRow(ctx, query, name, secret).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
