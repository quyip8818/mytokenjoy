package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
	pkgrl "github.com/tokenjoy/backend/internal/pkg/ratelimit"
)

// RedisLimiter implements distributed rate limiting using Redis + Lua scripts.
type RedisLimiter struct {
	rdb              *goredis.Client
	tokenBucketSHA   string
	slidingWindowSHA string
	logger           *slog.Logger
}

// NewRedisLimiter creates a Limiter backed by Redis.
// Returns nil if redisURL is empty or Redis is unreachable.
func NewRedisLimiter(ctx context.Context, redisURL string, logger *slog.Logger) (*RedisLimiter, error) {
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("ratelimit: parse REDIS_URL: %w", err)
	}
	rdb := goredis.NewClient(opt)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ratelimit: redis ping: %w", err)
	}

	// Load Lua scripts.
	tbSHA, err := rdb.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ratelimit: load token bucket script: %w", err)
	}
	swSHA, err := rdb.ScriptLoad(ctx, slidingWindowScript).Result()
	if err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ratelimit: load sliding window script: %w", err)
	}

	return &RedisLimiter{
		rdb:              rdb,
		tokenBucketSHA:   tbSHA,
		slidingWindowSHA: swSHA,
		logger:           logger,
	}, nil
}

func (l *RedisLimiter) AllowTokenBucket(ctx context.Context, key string, rate int, burst int) (pkgrl.Result, error) {
	now := time.Now().UnixMicro()
	// refillInterval = 1_000_000 / rate (microseconds per token)
	refillInterval := int64(1_000_000) / int64(rate)

	res, err := l.rdb.EvalSha(ctx, l.tokenBucketSHA, []string{key},
		burst,          // ARGV[1] capacity
		1,              // ARGV[2] refill_rate (tokens per refill)
		refillInterval, // ARGV[3] refill_interval (microseconds)
		now,            // ARGV[4] now (microseconds)
	).Int64Slice()
	if err != nil {
		return pkgrl.Result{}, err
	}

	allowed := res[0] == 1
	remaining := res[1]
	resetAt := time.Now().Add(time.Duration(refillInterval) * time.Microsecond)

	return pkgrl.Result{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     int64(burst),
		ResetAt:   resetAt,
	}, nil
}

func (l *RedisLimiter) AllowSlidingWindow(ctx context.Context, key string, max int, windowSec int) (pkgrl.Result, error) {
	now := time.Now().Unix()

	res, err := l.rdb.EvalSha(ctx, l.slidingWindowSHA, []string{key},
		max,       // ARGV[1] max requests
		windowSec, // ARGV[2] window in seconds
		now,       // ARGV[3] current timestamp
	).Int64Slice()
	if err != nil {
		return pkgrl.Result{}, err
	}

	allowed := res[0] == 1
	remaining := res[1]
	resetAt := time.Now().Add(time.Duration(windowSec) * time.Second)

	return pkgrl.Result{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     int64(max),
		ResetAt:   resetAt,
	}, nil
}

func (l *RedisLimiter) Close() error {
	if l == nil || l.rdb == nil {
		return nil
	}
	return l.rdb.Close()
}

var _ pkgrl.Limiter = (*RedisLimiter)(nil)
