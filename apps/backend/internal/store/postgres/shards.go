package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func loadShard(ctx context.Context, exec dbQuerier, id string) (json.RawMessage, error) {
	var raw []byte
	err := exec.QueryRow(ctx, `SELECT data FROM domain_snapshot WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shard %s not found", id)
		}
		slog.WarnContext(ctx, "postgres shard load failed", "shard", id, "error", err)
		return nil, fmt.Errorf("load shard %s: %w", id, err)
	}
	return json.RawMessage(raw), nil
}

func upsertShard(ctx context.Context, exec dbQuerier, id string, raw json.RawMessage) error {
	_, err := exec.Exec(ctx, `
		INSERT INTO domain_snapshot (id, data, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()
	`, id, raw)
	if err != nil {
		return fmt.Errorf("persist shard %s: %w", id, err)
	}
	return nil
}

func seedShards(ctx context.Context, exec dbQuerier, snapshot store.Snapshot) error {
	payloads, err := store.SnapshotToShards(snapshot)
	if err != nil {
		return err
	}
	for _, id := range store.AllShardIDs() {
		raw, ok := payloads[id]
		if !ok {
			return fmt.Errorf("unknown shard %s", id)
		}
		if err := upsertShard(ctx, exec, id, raw); err != nil {
			return err
		}
	}
	return nil
}

func loadAllShards(ctx context.Context, exec dbQuerier) (map[string]json.RawMessage, error) {
	rows, err := exec.Query(ctx, `
		SELECT id, data FROM domain_snapshot WHERE id = ANY($1)
	`, store.AllShardIDs())
	if err != nil {
		return nil, fmt.Errorf("load shards: %w", err)
	}
	defer rows.Close()

	shards := make(map[string]json.RawMessage)
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, fmt.Errorf("scan shard: %w", err)
		}
		shards[id] = json.RawMessage(raw)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shards: %w", err)
	}
	return shards, nil
}

func shardsComplete(shards map[string]json.RawMessage) bool {
	for _, id := range store.AllShardIDs() {
		if _, ok := shards[id]; !ok {
			return false
		}
	}
	return true
}

type shardBackend interface {
	load(shardID string) (json.RawMessage, error)
	save(shardID string, raw json.RawMessage) error
}

type poolShardBackend struct {
	ctx  context.Context
	exec dbQuerier
}

func newPoolShardBackend(ctx context.Context, exec dbQuerier) *poolShardBackend {
	if ctx == nil {
		ctx = context.Background()
	}
	return &poolShardBackend{ctx: ctx, exec: exec}
}

func (b *poolShardBackend) load(shardID string) (json.RawMessage, error) {
	return loadShard(b.ctx, b.exec, shardID)
}

func (b *poolShardBackend) save(shardID string, raw json.RawMessage) error {
	return upsertShard(b.ctx, b.exec, shardID, raw)
}

type txShardCache struct {
	ctx    context.Context
	exec   dbQuerier
	mu     sync.Mutex
	shards map[string]json.RawMessage
	dirty  map[string]struct{}
}

func newTxShardCache(ctx context.Context, exec dbQuerier) *txShardCache {
	return &txShardCache{
		ctx:    ctx,
		exec:   exec,
		shards: make(map[string]json.RawMessage),
		dirty:  make(map[string]struct{}),
	}
}

func (c *txShardCache) load(shardID string) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if raw, ok := c.shards[shardID]; ok {
		return raw, nil
	}
	raw, err := loadShard(c.ctx, c.exec, shardID)
	if err != nil {
		return nil, err
	}
	c.shards[shardID] = raw
	return raw, nil
}

func (c *txShardCache) save(shardID string, raw json.RawMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shards[shardID] = raw
	c.dirty[shardID] = struct{}{}
	return nil
}

func (c *txShardCache) flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for shardID := range c.dirty {
		if err := upsertShard(ctx, c.exec, shardID, c.shards[shardID]); err != nil {
			return err
		}
	}
	return nil
}
