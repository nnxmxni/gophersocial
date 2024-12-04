package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
)

type Follow struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowStore struct {
	db *sql.DB
}

func (s *FollowStore) Follow(ctx context.Context, userID int64, followerID int64) error {

	query := `INSERT INTO followers(user_id, follower_id) VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			fmt.Print(pqErr.Code)
			if pqErr.Code == "23514" && pqErr.Constraint == "chk_user_not_self_follow" {
				return ErrSelfFollow
			}
			if pqErr.Code == "23505" && pqErr.Constraint == "followers_pkey" {
				return ErrDuplicateFollow
			}
		}

		return err
	}

	return nil
}

func (s *FollowStore) Unfollow(ctx context.Context, userID int64, followerID int64) error {

	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)

	if err != nil {
		return err
	}

	return nil
}
