package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var allowScript = redis.NewScript(`
local c = redis.call('INCR', KEYS[1])
if c == 1 then
redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return c`)

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

	count, err := allowScript.Run(ctx, l.rdb, []string{key}, int(l.window.Seconds())).Int64()
	if err != nil {
		return false, err
	}
	return count <= l.limit, nil
}
