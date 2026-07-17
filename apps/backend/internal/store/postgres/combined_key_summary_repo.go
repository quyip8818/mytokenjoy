package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

type combinedKeySummaryRepo struct {
	db dbQuerier
}

func newCombinedKeySummaryRepo(db dbQuerier) *combinedKeySummaryRepo {
	return &combinedKeySummaryRepo{db: db}
}

var _ store.CombinedKeySummaryRepository = (*combinedKeySummaryRepo)(nil)

func (r *combinedKeySummaryRepo) UpdateBatch(ctx context.Context, updates []store.CombinedKeySummaryUpdate) ([]store.CombinedKeySummary, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	ids := make([]string, len(updates))
	remains := make([]float64, len(updates))
	for i, u := range updates {
		ids[i] = u.PlatformKeyID.String()
		remains[i] = u.Remain
	}
	rows, err := r.db.Query(ctx, `
		WITH input AS (
			SELECT unnest($2::text[]) AS platform_key_id,
			       unnest($3::numeric[]) AS remain
		),
		updated AS (
			UPDATE platform_keys pk
			SET combined_key_remain = input.remain,
			    combined_key_remain_at = NOW(),
			    combined_key_remain_version = pk.combined_key_remain_version + 1
			FROM input
			WHERE pk.company_id = $1
			  AND pk.id = input.platform_key_id
			RETURNING pk.id, pk.key_hash, pk.combined_key_remain, pk.combined_key_remain_at, pk.combined_key_remain_version
		)
		SELECT id, key_hash, combined_key_remain, combined_key_remain_at, combined_key_remain_version
		FROM updated
	`, companyID, ids, remains)
	if err != nil {
		return nil, fmt.Errorf("update combined key summaries: %w", err)
	}
	defer rows.Close()

	out := make([]store.CombinedKeySummary, 0, len(updates))
	for rows.Next() {
		var item store.CombinedKeySummary
		var updatedAt *time.Time
		if err := rows.Scan(&item.PlatformKeyID, &item.KeyHash, &item.Remain, &updatedAt, &item.Version); err != nil {
			return nil, err
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *combinedKeySummaryRepo) DecrementBatch(ctx context.Context, decrements map[uuid.UUID]float64) ([]store.CombinedKeySummary, error) {
	if len(decrements) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	ids := make([]uuid.UUID, 0, len(decrements))
	deltas := make([]float64, 0, len(decrements))
	for id, delta := range decrements {
		ids = append(ids, id)
		deltas = append(deltas, delta)
	}
	rows, err := r.db.Query(ctx, `
		WITH input AS (
			SELECT unnest($2::uuid[]) AS platform_key_id,
			       unnest($3::numeric[]) AS delta
		),
		updated AS (
			UPDATE platform_keys pk
			SET combined_key_remain = GREATEST(pk.combined_key_remain - input.delta, 0),
			    combined_key_remain_at = NOW(),
			    combined_key_remain_version = pk.combined_key_remain_version + 1
			FROM input
			WHERE pk.company_id = $1
			  AND pk.id = input.platform_key_id
			  AND pk.combined_key_remain IS NOT NULL
			RETURNING pk.id, pk.key_hash, pk.combined_key_remain, pk.combined_key_remain_at, pk.combined_key_remain_version
		)
		SELECT id, key_hash, combined_key_remain, combined_key_remain_at, combined_key_remain_version
		FROM updated
	`, companyID, ids, deltas)
	if err != nil {
		return nil, fmt.Errorf("decrement combined key summaries: %w", err)
	}
	defer rows.Close()

	out := make([]store.CombinedKeySummary, 0, len(decrements))
	for rows.Next() {
		var item store.CombinedKeySummary
		var updatedAt *time.Time
		if err := rows.Scan(&item.PlatformKeyID, &item.KeyHash, &item.Remain, &updatedAt, &item.Version); err != nil {
			return nil, err
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *combinedKeySummaryRepo) ListByPlatformKeyIDs(ctx context.Context, keyIDs []uuid.UUID) ([]store.CombinedKeySummary, error) {
	if len(keyIDs) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, key_hash, combined_key_remain, combined_key_remain_at, combined_key_remain_version
		FROM platform_keys
		WHERE company_id = $1 AND id = ANY($2)
	`, companyID, keyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]store.CombinedKeySummary, 0, len(keyIDs))
	for rows.Next() {
		var item store.CombinedKeySummary
		var remain *float64
		var updatedAt *time.Time
		if err := rows.Scan(&item.PlatformKeyID, &item.KeyHash, &remain, &updatedAt, &item.Version); err != nil {
			return nil, err
		}
		if remain != nil {
			item.Remain = *remain
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *combinedKeySummaryRepo) LockPlatformKeysForUpdate(ctx context.Context, keyIDs []uuid.UUID) error {
	if len(keyIDs) == 0 {
		return nil
	}
	companyID := store.CompanyID(ctx)
	// Lock rows in stable order to prevent deadlocks.
	_, err := r.db.Exec(ctx, `
		SELECT 1 FROM platform_keys
		WHERE company_id = $1 AND id = ANY($2)
		ORDER BY id
		FOR UPDATE
	`, companyID, keyIDs)
	return err
}
