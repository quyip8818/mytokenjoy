package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type tenantBackgroundStateRepo struct {
	db dbQuerier
}

func newTenantBackgroundStateRepo(db dbQuerier) store.TenantBackgroundStateRepository {
	return &tenantBackgroundStateRepo{db: db}
}

func (r *tenantBackgroundStateRepo) EnsureRow(ctx context.Context, companyID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO tenant_background_state (company_id)
		VALUES ($1)
		ON CONFLICT (company_id) DO NOTHING
	`, companyID)
	return err
}

func (r *tenantBackgroundStateRepo) Get(ctx context.Context, companyID int64) (*store.TenantBackgroundState, error) {
	row := r.db.QueryRow(ctx, `
		SELECT company_id, next_org_sync_at, last_org_sync_at,
		       last_rebalanced_period, last_budget_reconcile_at,
		       last_dashboard_reconcile_at, updated_at
		FROM tenant_background_state
		WHERE company_id = $1
	`, companyID)

	var state store.TenantBackgroundState
	err := row.Scan(
		&state.CompanyID,
		&state.NextOrgSyncAt,
		&state.LastOrgSyncAt,
		&state.LastRebalancedPeriod,
		&state.LastBudgetReconcileAt,
		&state.LastDashboardReconcileAt,
		&state.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *tenantBackgroundStateRepo) UpsertOrgSchedule(ctx context.Context, companyID int64, nextOrgSyncAt time.Time, lastOrgSyncAt *time.Time) error {
	if err := r.EnsureRow(ctx, companyID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE tenant_background_state
		SET next_org_sync_at = $2,
		    last_org_sync_at = COALESCE($3, last_org_sync_at),
		    updated_at = NOW()
		WHERE company_id = $1
	`, companyID, nextOrgSyncAt.UTC(), lastOrgSyncAt)
	return err
}

func (r *tenantBackgroundStateRepo) SetLastRebalancedPeriod(ctx context.Context, companyID int64, period string) error {
	if err := r.EnsureRow(ctx, companyID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE tenant_background_state
		SET last_rebalanced_period = $2,
		    updated_at = NOW()
		WHERE company_id = $1
	`, companyID, period)
	return err
}

func (r *tenantBackgroundStateRepo) SetLastBudgetReconcileAt(ctx context.Context, companyID int64, at time.Time) error {
	if err := r.EnsureRow(ctx, companyID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE tenant_background_state
		SET last_budget_reconcile_at = $2,
		    updated_at = NOW()
		WHERE company_id = $1
	`, companyID, at.UTC())
	return err
}

func (r *tenantBackgroundStateRepo) SetLastDashboardReconcileAt(ctx context.Context, companyID int64, at time.Time) error {
	if err := r.EnsureRow(ctx, companyID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE tenant_background_state
		SET last_dashboard_reconcile_at = $2,
		    updated_at = NOW()
		WHERE company_id = $1
	`, companyID, at.UTC())
	return err
}
