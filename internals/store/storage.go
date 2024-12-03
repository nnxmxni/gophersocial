package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetPostByID(context.Context, int64) (*Post, error)
		Update(context.Context, *Post) error
		Delete(context.Context, int64) error
	}
	Users interface {
		Create(context.Context, *User) error
		GetUserByID(context.Context, int64) (*User, error)
	}
	Comment interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:   &PostStore{db: db},
		Users:   &UserStore{db: db},
		Comment: &CommentStore{db},
	}
}
