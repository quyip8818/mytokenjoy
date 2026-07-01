package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) ProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, provider, name, key_prefix, secret_key, relay_channel_id, status,
			balance, last_used, rotate_enabled, created_at
		FROM provider_keys ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.ProviderKey, 0)
	for rows.Next() {
		var item types.ProviderKey
		var lastUsed *time.Time
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Provider, &item.Name, &item.KeyPrefix, &item.SecretKey,
			&item.RelayChannelID, &item.Status, &item.Balance, &lastUsed,
			&item.RotateEnabled, &createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatDateOnly(createdAt)
		if lastUsed != nil {
			s := formatSyncLogTime(*lastUsed)
			item.LastUsed = &s
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneProviderKeys(items), nil
}

func (r *pgKeysRepo) SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error {
	cloned := store.CloneProviderKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
		createdAt, err := parseAPITime(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var lastUsed *time.Time
		if key.LastUsed != nil {
			t, err := parseAPITime(*key.LastUsed)
			if err != nil {
				return err
			}
			lastUsed = &t
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, relay_channel_id, status,
				balance, last_used, rotate_enabled, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			ON CONFLICT (id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				secret_key = EXCLUDED.secret_key,
				relay_channel_id = EXCLUDED.relay_channel_id,
				status = EXCLUDED.status,
				balance = EXCLUDED.balance,
				last_used = EXCLUDED.last_used,
				rotate_enabled = EXCLUDED.rotate_enabled,
				updated_at = NOW()
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, key.SecretKey, key.RelayChannelID,
			key.Status, key.Balance, lastUsed, key.RotateEnabled, createdAt); err != nil {
			return fmt.Errorf("upsert provider key %s: %w", key.ID, err)
		}
	}
	return pruneByID(ctx, r.db, "provider_keys", ids)
}
