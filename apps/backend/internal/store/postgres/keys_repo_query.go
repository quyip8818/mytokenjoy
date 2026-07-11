package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) SumMemberKeyUsed(ctx context.Context, memberID string, clk clock.Clock) (float64, error) {
	companyID := store.CompanyID(ctx)
	var orgPeriod string
	err := r.db.QueryRow(ctx, `
		SELECT o.period
		FROM members m
		JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
		WHERE m.company_id = $1 AND m.id = $2
	`, companyID, memberID).Scan(&orgPeriod)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}
	periodKey := pkgbudget.OpenSnapshotKey(orgPeriod, clk).String()
	var total float64
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(consumed, 0) FROM budget_snapshots
		WHERE company_id = $1 AND axis_kind = $2 AND axis_id = $3 AND period_key = $4
	`, companyID, store.SnapshotAxisMember, memberID, periodKey).Scan(&total)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return total, err
}

func (r *pgKeysRepo) ListActiveMemberKeys(ctx context.Context, memberID string) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, platformKeySelect+`
		WHERE company_id = $1 AND member_id = $2 AND budget_group_id IS NULL AND status = 'active'
		ORDER BY id
	`, companyID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		item, err := scanPlatformKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
