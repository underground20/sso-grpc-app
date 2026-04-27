package storage

import (
	"app/internal/infrastructure/db"
	"app/internal/models"
	"context"
	"fmt"
	"strings"
)

type RoleStorage struct {
	db *db.Database
}

func NewRoleStorage(db *db.Database) RoleStorage {
	return RoleStorage{db: db}
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

func (s RoleStorage) CreateRole(ctx context.Context, name string, permissions []string) error {
	transaction, err := s.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			transaction.Rollback(ctx)
			panic(r)
		} else if err != nil {
			transaction.Rollback(ctx)
		}
	}()

	var roleID int
	err = transaction.QueryRow(ctx, `INSERT INTO roles(name) VALUES ($1) RETURNING id`, name).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to insert role: %w", err)
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
			return fmt.Errorf("failed to insert permissions for role %d: %w", roleID, err)
		}
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
