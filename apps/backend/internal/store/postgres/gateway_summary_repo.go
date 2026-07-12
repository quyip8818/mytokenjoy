package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type gatewaySoftSummaryRepo struct {
	db dbQuerier
}

func newGatewaySoftSummaryRepo(db dbQuerier) *gatewaySoftSummaryRepo {
	return &gatewaySoftSummaryRepo{db: db}
}

var _ store.GatewaySoftSummaryRepository = (*gatewaySoftSummaryRepo)(nil)

func (r *gatewaySoftSummaryRepo) UpdateBatch(ctx context.Context, updates []store.GatewaySoftSummaryUpdate) ([]store.GatewaySoftSummary, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	ids := make([]string, len(updates))
	remains := make([]float64, len(updates))
	for i, u := range updates {
		ids[i] = u.PlatformKeyID
		remains[i] = u.SoftRemain
	}
	rows, err := r.db.Query(ctx, `
		WITH input AS (
			SELECT unnest($2::text[]) AS platform_key_id,
			       unnest($3::numeric[]) AS soft_remain
		),
		updated AS (
			UPDATE platform_keys pk
			SET gateway_soft_remain = input.soft_remain,
			    gateway_soft_at = NOW(),
			    gateway_soft_version = pk.gateway_soft_version + 1
			FROM input
			WHERE pk.company_id = $1
			  AND pk.id = input.platform_key_id
			RETURNING pk.id, pk.key_hash, pk.gateway_soft_remain, pk.gateway_soft_at, pk.gateway_soft_version
		)
		SELECT id, key_hash, gateway_soft_remain, gateway_soft_at, gateway_soft_version
		FROM updated
	`, companyID, ids, remains)
	if err != nil {
		return nil, fmt.Errorf("update gateway soft summaries: %w", err)
	}
	defer rows.Close()

	out := make([]store.GatewaySoftSummary, 0, len(updates))
	for rows.Next() {
		var item store.GatewaySoftSummary
		var updatedAt *time.Time
		if err := rows.Scan(&item.PlatformKeyID, &item.KeyHash, &item.SoftRemain, &updatedAt, &item.Version); err != nil {
			return nil, err
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *gatewaySoftSummaryRepo) ListByPlatformKeyIDs(ctx context.Context, keyIDs []string) ([]store.GatewaySoftSummary, error) {
	if len(keyIDs) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, key_hash, gateway_soft_remain, gateway_soft_at, gateway_soft_version
		FROM platform_keys
		WHERE company_id = $1 AND id = ANY($2)
	`, companyID, keyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]store.GatewaySoftSummary, 0, len(keyIDs))
	for rows.Next() {
		var item store.GatewaySoftSummary
		var remain *float64
		var updatedAt *time.Time
		if err := rows.Scan(&item.PlatformKeyID, &item.KeyHash, &remain, &updatedAt, &item.Version); err != nil {
			return nil, err
		}
		if remain != nil {
			item.SoftRemain = *remain
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
