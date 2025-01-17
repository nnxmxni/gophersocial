package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nnxmxni/gophersocial/internals/store"
	"time"
)

type UserStore struct {
	rdb *redis.Client
}

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user *store.User

	if data != "" {
		err = json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {

	cacheKey := fmt.Sprintf("user-%v", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, data, time.Minute).Err()
}
