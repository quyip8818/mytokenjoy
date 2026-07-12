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
SELECT
	c.id AS company_id,
	c.status AS company_status,
	c.wallet_remain,
	c.newapi_wallet_user_id,
	pk.id AS platform_key_id,
	pk.status AS key_status,
	EXISTS (
		SELECT 1 FROM model_allowlist ma
		WHERE ma.company_id = pk.company_id
		  AND ma.owner_type = 'platform_key'
		  AND ma.owner_id = pk.id
	) AS has_allowlist,
	COALESCE((
		SELECT array_agg(DISTINCT mdl.type ORDER BY mdl.type)
		FROM model_allowlist ma
		JOIN models mdl ON mdl.model_id = ma.model_id
		WHERE ma.company_id = pk.company_id
		  AND ma.owner_type = 'platform_key'
		  AND ma.owner_id = pk.id
		  AND mdl.enabled = TRUE
	), '{}') AS allowlist_types,
	pk.gateway_soft_remain,
	pk.gateway_soft_at,
	pk.gateway_soft_version
FROM platform_keys pk
JOIN companies c ON c.id = pk.company_id
WHERE pk.key_hash = $1
`

func (r *gatewayPrecheckRepo) LoadPrecheckContext(ctx context.Context, keyHash string) (*store.PrecheckContextRow, error) {
	row := r.db.QueryRow(ctx, loadPrecheckContextSQL, keyHash)

	var out store.PrecheckContextRow
	var newAPIWalletUserID *int64
	err := row.Scan(
		&out.CompanyID,
		&out.CompanyStatus,
		&out.WalletRemain,
		&newAPIWalletUserID,
		&out.PlatformKeyID,
		&out.KeyStatus,
		&out.HasAllowlist,
		&out.AllowlistTypes,
		&out.GatewaySoftRemain,
		&out.GatewaySoftAt,
		&out.GatewaySoftVersion,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out.NewAPIWalletUserID = newAPIWalletUserID
	if out.AllowlistTypes == nil {
		out.AllowlistTypes = []string{}
	}
	return &out, nil
}
