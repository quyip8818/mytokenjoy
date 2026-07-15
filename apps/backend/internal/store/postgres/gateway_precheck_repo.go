package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type gatewayPrecheckRepo struct {
	db dbQuerier
}

func newGatewayPrecheckRepo(db dbQuerier) *gatewayPrecheckRepo {
	return &gatewayPrecheckRepo{db: db}
}

var _ store.GatewayPrecheckRepository = (*gatewayPrecheckRepo)(nil)

const loadPrecheckContextSQL = `
WITH pk_ctx AS (
	SELECT pk.id, pk.company_id, pk.status, pk.expires_at, pk.combined_key_remain, pk.combined_key_remain_version
	FROM platform_keys pk
	WHERE pk.key_hash = $1
),
allowlist AS (
	SELECT
		ma.owner_id AS platform_key_id,
		COUNT(*) > 0 AS has_allowlist,
		COALESCE(
			array_agg(DISTINCT mdl.type ORDER BY mdl.type) FILTER (WHERE mdl.enabled = TRUE),
			'{}'
		) AS allowlist_types
	FROM model_allowlist ma
	JOIN models mdl ON mdl.model_id = ma.model_id
	JOIN pk_ctx ON pk_ctx.id = ma.owner_id AND ma.company_id = pk_ctx.company_id
	WHERE ma.owner_type = 'platform_key'
	GROUP BY ma.owner_id
)
SELECT
	c.id AS company_id,
	c.status AS company_status,
	c.wallet_remain,
	pk_ctx.id AS platform_key_id,
	pk_ctx.status AS key_status,
	pk_ctx.expires_at AS key_expires_at,
	COALESCE(a.has_allowlist, FALSE) AS has_allowlist,
	COALESCE(a.allowlist_types, '{}') AS allowlist_types,
	pk_ctx.combined_key_remain,
	pk_ctx.combined_key_remain_version
FROM pk_ctx
JOIN companies c ON c.id = pk_ctx.company_id
LEFT JOIN allowlist a ON a.platform_key_id = pk_ctx.id
`

func (r *gatewayPrecheckRepo) LoadPrecheckContext(ctx context.Context, keyHash string) (*store.PrecheckContextRow, error) {
	row := r.db.QueryRow(ctx, loadPrecheckContextSQL, keyHash)

	var out store.PrecheckContextRow
	err := row.Scan(
		&out.CompanyID,
		&out.CompanyStatus,
		&out.WalletRemain,
		&out.PlatformKeyID,
		&out.KeyStatus,
		&out.KeyExpiresAt,
		&out.HasAllowlist,
		&out.AllowlistTypes,
		&out.CombinedKeyRemain,
		&out.CombinedKeyRemainVersion,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if out.AllowlistTypes == nil {
		out.AllowlistTypes = []string{}
	}
	return &out, nil
}
