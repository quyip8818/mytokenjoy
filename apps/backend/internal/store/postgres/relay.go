package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/store"
)

type dbQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type relayRepo struct {
	db dbQuerier
}

func newRelayRepo(db dbQuerier) *relayRepo {
	return &relayRepo{db: db}
}

func scanMapping(row pgx.Row) (store.RelayMapping, error) {
	var m store.RelayMapping
	var memberID, budgetGroupID *string
	var tokenID, remainQuota *int64
	var syncedAt *time.Time
	err := row.Scan(
		&m.CompanyID, &m.PlatformKeyID, &tokenID, &memberID, &m.DepartmentID, &budgetGroupID,
		&m.RelayGroup, &m.SyncStatus, &syncedAt, &remainQuota,
	)
	if err != nil {
		return store.RelayMapping{}, err
	}
	m.NewAPITokenID = tokenID
	m.MemberID = memberID
	m.BudgetGroupID = budgetGroupID
	m.SyncedAt = syncedAt
	m.RelayRemainQuota = remainQuota
	return m, nil
}

const mappingSelect = `
	SELECT company_id, platform_key_id, newapi_token_id, member_id, department_id, budget_group_id,
	       relay_group, sync_status, synced_at, relay_remain_quota
	FROM relay_mappings
`

var _ store.RelayRepository = (*relayRepo)(nil)
