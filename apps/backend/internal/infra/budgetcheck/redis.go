package budgetcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type redisStore struct {
	rdb *goredis.Client
	ttl time.Duration
}

func newRedisStore(ctx context.Context, redisURL string, ttl time.Duration) (*redisStore, error) {
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	rdb := goredis.NewClient(opt)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &redisStore{rdb: rdb, ttl: ttl}, nil
}

func (g *redisStore) Enabled() bool { return g != nil && g.rdb != nil }

func (g *redisStore) Get(ctx context.Context, companyID int64, keyHash string) (Entry, bool, error) {
	raw, err := g.rdb.Get(ctx, Key(companyID, keyHash)).Bytes()
	if err == goredis.Nil {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	var entry Entry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return Entry{}, false, nil
	}
	return entry, true, nil
}

func (g *redisStore) Set(ctx context.Context, companyID int64, keyHash string, entry Entry) error {
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = time.Now().UTC()
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return g.rdb.Set(ctx, Key(companyID, keyHash), raw, g.ttl).Err()
}

func (g *redisStore) Close() error {
	if g == nil || g.rdb == nil {
		return nil
	}
	return g.rdb.Close()
}
