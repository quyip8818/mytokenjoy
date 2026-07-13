package store

import (
	"context"
	"time"
)

type TenantBackgroundState struct {
	CompanyID                int64
	NextOrgSyncAt            *time.Time
	LastOrgSyncAt            *time.Time
	LastRebalancedPeriod     string
	LastBudgetReconcileAt    *time.Time
	LastDashboardReconcileAt *time.Time
	UpdatedAt                time.Time
}

type TenantBackgroundStateRepository interface {
	EnsureRow(ctx context.Context, companyID int64) error
	Get(ctx context.Context, companyID int64) (*TenantBackgroundState, error)
	UpsertOrgSchedule(ctx context.Context, companyID int64, nextOrgSyncAt time.Time, lastOrgSyncAt *time.Time) error
	SetLastRebalancedPeriod(ctx context.Context, companyID int64, period string) error
	SetLastBudgetReconcileAt(ctx context.Context, companyID int64, at time.Time) error
	SetLastDashboardReconcileAt(ctx context.Context, companyID int64, at time.Time) error
}
