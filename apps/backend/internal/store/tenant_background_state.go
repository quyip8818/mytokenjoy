package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TenantBackgroundState struct {
	CompanyID                uuid.UUID
	NextOrgSyncAt            *time.Time
	LastOrgSyncAt            *time.Time
	LastRebalancedPeriod     string
	LastBudgetReconcileAt    *time.Time
	LastDashboardReconcileAt *time.Time
	UpdatedAt                time.Time
}

type TenantBackgroundStateRepository interface {
	EnsureRow(ctx context.Context, companyID uuid.UUID) error
	Get(ctx context.Context, companyID uuid.UUID) (*TenantBackgroundState, error)
	UpsertOrgSchedule(ctx context.Context, companyID uuid.UUID, nextOrgSyncAt time.Time, lastOrgSyncAt *time.Time) error
	SetLastRebalancedPeriod(ctx context.Context, companyID uuid.UUID, period string) error
	SetLastBudgetReconcileAt(ctx context.Context, companyID uuid.UUID, at time.Time) error
	SetLastDashboardReconcileAt(ctx context.Context, companyID uuid.UUID, at time.Time) error
}
