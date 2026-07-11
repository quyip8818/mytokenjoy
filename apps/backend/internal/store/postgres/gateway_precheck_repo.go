package postgres

import (
	"context"
	"errors"
	"time"

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
WITH base AS (
	SELECT
		c.id AS company_id,
		c.status AS company_status,
		c.balance_point,
		c.newapi_wallet_user_id,
		pk.id AS platform_key_id,
		pk.status AS key_status,
		pk.budget AS key_budget,
		pk.member_id,
		pk.budget_group_id,
		COALESCE(
			NULLIF(m.department_id, ''),
			CASE WHEN pkm.newapi_group LIKE 'dept-%' THEN substring(pkm.newapi_group FROM 6) ELSE '' END
		) AS department_id,
		(on_node.id IS NOT NULL) AS dept_found,
		COALESCE(on_node.budget, 0) AS dept_budget,
		COALESCE(on_node.period, 'monthly') AS dept_period,
		m.personal_budget AS member_cap,
		(m.id IS NOT NULL) AS member_found,
		COALESCE(bg.budget, 0) AS group_budget,
		CASE
			WHEN on_node.period IS NOT NULL AND on_node.period <> 'monthly' THEN on_node.period
			ELSE to_char($2::timestamptz AT TIME ZONE 'UTC', 'YYYY-MM')
		END AS period_key
	FROM platform_keys pk
	JOIN platform_key_mappings pkm
		ON pkm.company_id = pk.company_id AND pkm.platform_key_id = pk.id
	JOIN companies c ON c.id = pk.company_id
	LEFT JOIN members m ON m.company_id = pk.company_id AND m.id = pk.member_id
	LEFT JOIN org_nodes on_node
		ON on_node.company_id = pk.company_id
		AND on_node.id = COALESCE(
			NULLIF(m.department_id, ''),
			CASE WHEN pkm.newapi_group LIKE 'dept-%' THEN substring(pkm.newapi_group FROM 6) ELSE NULL END
		)
	LEFT JOIN budget_groups bg
		ON bg.company_id = pk.company_id AND bg.id = pk.budget_group_id
	WHERE pk.key_hash = $1
)
SELECT
	b.company_id,
	b.company_status,
	b.balance_point,
	b.newapi_wallet_user_id,
	b.platform_key_id,
	b.key_status,
	b.key_budget,
	b.department_id,
	b.dept_found,
	b.dept_budget,
	b.period_key,
	b.member_id,
	b.member_found,
	b.member_cap,
	b.budget_group_id,
	b.group_budget,
	COALESCE(key_snap.consumed, 0) AS key_consumed,
	COALESCE(dept_snap.consumed, 0) AS dept_consumed,
	COALESCE(member_snap.consumed, 0) AS member_consumed,
	COALESCE(bg_snap.consumed, 0) AS group_consumed,
	EXISTS (
		SELECT 1 FROM model_allowlist ma
		WHERE ma.company_id = b.company_id
			AND ma.owner_type = 'platform_key'
			AND ma.owner_id = b.platform_key_id
	) AS has_allowlist,
	COALESCE((
		SELECT array_agg(DISTINCT mdl.type ORDER BY mdl.type)
		FROM model_allowlist ma
		JOIN models mdl ON mdl.model_id = ma.model_id
		WHERE ma.company_id = b.company_id
			AND ma.owner_type = 'platform_key'
			AND ma.owner_id = b.platform_key_id
			AND mdl.enabled = TRUE
	), '{}') AS allowlist_types
FROM base b
LEFT JOIN budget_snapshots key_snap
	ON key_snap.company_id = b.company_id
	AND key_snap.axis_kind = 'platform_key'
	AND key_snap.axis_id = b.platform_key_id
	AND key_snap.period_key = b.period_key
LEFT JOIN budget_snapshots dept_snap
	ON dept_snap.company_id = b.company_id
	AND dept_snap.axis_kind = 'org_node'
	AND dept_snap.axis_id = b.department_id
	AND dept_snap.period_key = b.period_key
LEFT JOIN budget_snapshots member_snap
	ON member_snap.company_id = b.company_id
	AND member_snap.axis_kind = 'member'
	AND member_snap.axis_id = b.member_id
	AND member_snap.period_key = b.period_key
LEFT JOIN budget_snapshots bg_snap
	ON bg_snap.company_id = b.company_id
	AND bg_snap.axis_kind = 'budget_group'
	AND bg_snap.axis_id = b.budget_group_id
	AND bg_snap.period_key = b.period_key
`

func (r *gatewayPrecheckRepo) LoadPrecheckContext(ctx context.Context, keyHash string, at time.Time) (*store.PrecheckContextRow, error) {
	row := r.db.QueryRow(ctx, loadPrecheckContextSQL, keyHash, at.UTC())

	var out store.PrecheckContextRow
	var newAPIWalletUserID *int64
	var memberID, budgetGroupID *string
	var departmentID string

	err := row.Scan(
		&out.CompanyID,
		&out.CompanyStatus,
		&out.BalancePoint,
		&newAPIWalletUserID,
		&out.PlatformKeyID,
		&out.KeyStatus,
		&out.KeyBudget,
		&departmentID,
		&out.DeptFound,
		&out.DeptBudget,
		&out.PeriodKey,
		&memberID,
		&out.MemberFound,
		&out.MemberCap,
		&budgetGroupID,
		&out.GroupBudget,
		&out.KeyConsumed,
		&out.DeptConsumed,
		&out.MemberConsumed,
		&out.GroupConsumed,
		&out.HasAllowlist,
		&out.AllowlistTypes,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	out.NewAPIWalletUserID = newAPIWalletUserID
	out.DepartmentID = departmentID
	out.MemberID = memberID
	out.BudgetGroupID = budgetGroupID
	if out.AllowlistTypes == nil {
		out.AllowlistTypes = []string{}
	}
	return &out, nil
}
