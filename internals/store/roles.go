package store

import (
	"context"
	"database/sql"
)

type Roles struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Roles, error) {

	query := `SELECT id, name, level, description FROM roles WHERE name = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	roles := &Roles{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		roleName,
	).Scan(
		&roles.ID,
		&roles.Name,
		&roles.Level,
		&roles.Description,
	)

	if err != nil {
		return nil, err
	}

	return roles, nil
}
