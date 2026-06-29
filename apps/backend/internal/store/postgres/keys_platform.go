package postgres

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) PlatformKeys() []types.PlatformKey {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, key_prefix, full_key, member_id, member_name, app_name,
			budget_group_id, budget_group_name, status, quota, used, created_at, expires_at
		FROM platform_keys ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		var item types.PlatformKey
		var createdAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Name, &item.KeyPrefix, &item.FullKey, &item.MemberID, &item.MemberName,
			&item.AppName, &item.BudgetGroupID, &item.BudgetGroupName, &item.Status,
			&item.Quota, &item.Used, &createdAt, &expiresAt,
		); err != nil {
			return nil
		}
		item.CreatedAt = formatDateOnly(createdAt)
		if expiresAt != nil {
			s := formatDateOnly(*expiresAt)
			item.ExpiresAt = &s
		}
		modelRows, err := r.db.Query(r.ctx, `
			SELECT model_name FROM platform_key_models WHERE platform_key_id = $1 ORDER BY model_name
		`, item.ID)
		if err == nil {
			for modelRows.Next() {
				var modelName string
				if err := modelRows.Scan(&modelName); err == nil {
					item.ModelWhitelist = append(item.ModelWhitelist, modelName)
				}
			}
			modelRows.Close()
		}
		items = append(items, item)
	}
	return store.ClonePlatformKeys(items)
}

func (r *pgKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	cloned := store.ClonePlatformKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
		createdAt, err := parseAPITime(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := parseAPITime(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO platform_keys (
				id, name, key_prefix, full_key, member_id, member_name, app_name,
				budget_group_id, budget_group_name, status, quota, used, created_at, expires_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				full_key = EXCLUDED.full_key,
				member_id = EXCLUDED.member_id,
				member_name = EXCLUDED.member_name,
				app_name = EXCLUDED.app_name,
				budget_group_id = EXCLUDED.budget_group_id,
				budget_group_name = EXCLUDED.budget_group_name,
				status = EXCLUDED.status,
				quota = EXCLUDED.quota,
				used = EXCLUDED.used,
				expires_at = EXCLUDED.expires_at,
				updated_at = NOW()
		`, key.ID, key.Name, key.KeyPrefix, key.FullKey, key.MemberID, key.MemberName,
			key.AppName, key.BudgetGroupID, key.BudgetGroupName, key.Status,
			key.Quota, key.Used, createdAt, expiresAt); err != nil {
			return fmt.Errorf("upsert platform key %s: %w", key.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM platform_key_models WHERE platform_key_id = $1`, key.ID); err != nil {
			return err
		}
		for _, modelName := range key.ModelWhitelist {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO platform_key_models (platform_key_id, model_name) VALUES ($1, $2)
			`, key.ID, modelName); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM platform_key_models`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM platform_keys`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "platform_key_models", "platform_key_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "platform_keys", ids)
}
