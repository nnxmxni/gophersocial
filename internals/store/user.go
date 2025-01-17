package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("the email already exists")
)

type User struct {
	ID              int64        `json:"id"`
	Email           string       `json:"email"`
	Password        password     `json:"-"`
	EmailVerifiedAt sql.NullTime `json:"email_verified_at"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Role            Roles        `json:"roles"`
}

type password struct {
	Text *string
	Hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.Text = &text
	p.Hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	if err := bcrypt.CompareHashAndPassword(p.Hash, []byte(text)); err != nil {
		return err
	}

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) create(ctx context.Context, tx *sql.Tx, user *User) error {

	query := `
		WITH role AS (
		    SELECT id, level, description
		    FROM roles
		    WHERE name = $3
		)
		INSERT INTO users (email, password, role_id) 
		VALUES (LOWER($1), $2, (SELECT id FROM role))
		RETURNING id, email_verified_at, created_at, updated_at,
		    (SELECT id FROM role),
		    (SELECT level FROM role),
		    (SELECT description FROM role)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Password.Hash,
		user.Role.Name,
	).Scan(
		&user.ID,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			fmt.Println(pqErr.Code)
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {

	query := `
		UPDATE users
		SET email = $1, email_verified_at = $2 
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		user.Email,
		user.EmailVerifiedAt,
		user.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT u.id, u.email, u.email_verified_at, u.created_at, u.updated_at, role_id, r.name, r.description, r.level
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
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

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {

	query := `
		SELECT u.id, u.email, u.password, u.email_verified_at, u.created_at, u.updated_at, role_id, r.name, r.description, r.level
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE email = $1 AND email_verified_at IS NOT NULL 
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password.Hash,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, err
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {

	query := `INSERT INTO one_time_passwords (token, user_id, expired_at) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		userID,
		time.Now().Add(exp),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		// find the user the token belongs to
		user, err := s.getUserFromInvitation(ctx, tx, token)

		//update the user
		user.EmailVerifiedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}

		if err = s.update(ctx, tx, user); err != nil {
			return err
		}

		//clear the invitation
		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {

	query := `
			SELECT u.id, u.email, u.email_verified_at, u.created_at
			FROM users u
			JOIN one_time_passwords otp ON u.id = otp.user_id
			WHERE otp.token = $1 AND otp.expired_at > $2
		`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
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

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {

	query := `
		DELETE FROM one_time_passwords WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return err
	}

	return nil
}
