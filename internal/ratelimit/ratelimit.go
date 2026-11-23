package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb    *redis.Client
	limit  int64
	window time.Duration
}

func New(rdb *redis.Client, limit int64, window time.Duration) *Limiter {
	return &Limiter{rdb: rdb, limit: limit, window: window}
}

func (l *Limiter) Allow(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("ratelimit:likes:%d", userID)

	count, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := l.rdb.Expire(ctx, key, l.window).Err(); err != nil {
			return false, err
		}
	}
	return count <= l.limit, nil
}
