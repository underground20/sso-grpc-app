package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type RoleStorage struct {
	db     *db.Database
	logger *slog.Logger
}

func NewRoleStorage(db *db.Database, logger *slog.Logger) RoleStorage {
	return RoleStorage{db: db, logger: logger}
}

func (s RoleStorage) GetRoles(ctx context.Context) ([]models.Role, error) {
	sql := `
		SELECT r.id, r.name, ARRAY_AGG(rp.permission) as permissions
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		GROUP BY r.id, r.name
		ORDER BY r.name
	`

	rows, err := s.db.Conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var id int
		var name string
		var permissions []string

		if err := rows.Scan(&id, &name, &permissions); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		roles = append(roles, models.Role{
			ID:          id,
			Name:        name,
			Permissions: permissions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return roles, nil
}

func (s RoleStorage) CreateRole(ctx context.Context, name string, permissions []string) (int, error) {
	transaction, err := s.db.Conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			s.logger.Error("failed to rollback transaction", slog.String("error", rollbackErr.Error()))
		}
	}()

	var roleID int
	err = transaction.QueryRow(ctx, `INSERT INTO roles(name) VALUES ($1) RETURNING id`, name).Scan(&roleID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert role: %w", err)
	}

	if len(permissions) > 0 {
		valueStrings := make([]string, 0, len(permissions))
		valueArgs := make([]interface{}, 0, len(permissions)*2)

		for i, perm := range permissions {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			valueArgs = append(valueArgs, roleID, perm)
		}

		stmt := fmt.Sprintf(`INSERT INTO role_permissions(role_id, permission) VALUES %s`, strings.Join(valueStrings, ","))
		_, err = transaction.Exec(ctx, stmt, valueArgs...)
		if err != nil {
			return 0, fmt.Errorf("failed to insert permissions for role %d: %w", roleID, err)
		}
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return roleID, nil
}

func (s RoleStorage) RolesExist(ctx context.Context, roleIDs []int64) (bool, error) {
	if len(roleIDs) == 0 {
		return true, nil // Если ролей нет, то и проверять нечего
	}

	var count int
	err := s.db.Conn.QueryRow(ctx,
		`SELECT count(*) FROM roles WHERE id = ANY($1)`,
		roleIDs,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check roles existence: %w", err)
	}

	return count == len(roleIDs), nil
}
