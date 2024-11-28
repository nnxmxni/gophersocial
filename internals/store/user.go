package store

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserStore struct {
	db *sql.DB
}

func (s UserStore) Create(ctx context.Context, user *User) error {

	query := `
		INSERT INTO users (email, password) 
		VALUES ($1, $2) RETURNING id, created_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s UserStore) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
