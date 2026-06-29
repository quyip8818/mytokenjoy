package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

func (s *Store) markDirty(shard string) {
	s.dirtyMu.Lock()
	defer s.dirtyMu.Unlock()
	if s.dirtyShards == nil {
		s.dirtyShards = make(map[string]struct{})
	}
	s.dirtyShards[shard] = struct{}{}
}

func (s *Store) clearDirty(shards ...string) {
	s.dirtyMu.Lock()
	defer s.dirtyMu.Unlock()
	for _, shard := range shards {
		delete(s.dirtyShards, shard)
	}
}

func (s *Store) persistShard(shard string) error {
	s.markDirty(shard)
	return s.flushShards(s.domainPersistCtx(), s.pool, shard)
}

func (s *Store) flushShards(ctx context.Context, exec dbQuerier, shards ...string) error {
	targets := shards
	if len(targets) == 0 {
		s.dirtyMu.Lock()
		targets = make([]string, 0, len(s.dirtyShards))
		for shard := range s.dirtyShards {
			targets = append(targets, shard)
		}
		s.dirtyMu.Unlock()
	}
	if len(targets) == 0 {
		return nil
	}

	payloads, err := store.SnapshotToShards(s.memory.Snapshot())
	if err != nil {
		return err
	}
	for _, shard := range targets {
		raw, ok := payloads[shard]
		if !ok {
			return fmt.Errorf("unknown shard %s", shard)
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO domain_snapshot (id, data, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()
		`, shard, raw); err != nil {
			return fmt.Errorf("persist shard %s: %w", shard, err)
		}
	}
	s.clearDirty(targets...)
	return nil
}

func (s *Store) loadShards(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := s.pool.Query(ctx, `
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

func (s *Store) seedDomainShards(ctx context.Context, seed store.Snapshot) error {
	s.memory = store.NewMemory(seed)
	for _, shard := range store.AllShardIDs() {
		s.markDirty(shard)
	}
	return s.flushShards(ctx, s.pool)
}

func shardsComplete(shards map[string]json.RawMessage) bool {
	for _, id := range store.AllShardIDs() {
		if _, ok := shards[id]; !ok {
			return false
		}
	}
	return true
}
