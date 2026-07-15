package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/store"
)

func InsertWalletSync(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := WalletSyncArgs{CompanyID: companyID}
	return insert(ctx, e, tx, args, nil)
}

func InsertRebalance(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64, axisKind, axisID string) error {
	args := RebalanceArgs{
		CompanyID: companyID,
		AxisKind:  axisKind,
		AxisID:    axisID,
	}
	return insert(ctx, e, tx, args, nil)
}

func InsertOverrun(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64, payload []byte) error {
	args := OverrunArgs{
		CompanyID: companyID,
		Payload:   payload,
	}
	return insert(ctx, e, tx, args, nil)
}

func InsertNewAPISync(ctx context.Context, e Enqueuer, tx store.Tx, args NewAPISyncArgs) error {
	return insert(ctx, e, tx, args, nil)
}

func InsertOrgSync(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64, scheduledAt *time.Time) error {
	args := OrgSyncArgs{CompanyID: companyID}
	var opts *river.InsertOpts
	if scheduledAt != nil {
		base := args.InsertOpts()
		base.ScheduledAt = scheduledAt.UTC()
		opts = &base
	}
	return insert(ctx, e, tx, args, opts)
}

func InsertBudgetReconcile(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := BudgetReconcileArgs{CompanyID: companyID}
	return insert(ctx, e, tx, args, nil)
}

func InsertDashboardProject(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := DashboardProjectArgs{CompanyID: companyID}
	return insert(ctx, e, tx, args, nil)
}

func InsertDashboardReconcile(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := DashboardReconcileArgs{CompanyID: companyID}
	return insert(ctx, e, tx, args, nil)
}

func InsertTenantWatchdog(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := TenantWatchdogArgs{}
	return insert(ctx, e, tx, args, nil)
}

func InsertNotificationDelivery(ctx context.Context, e Enqueuer, args NotificationDeliveryArgs) error {
	return insert(ctx, e, nil, args, nil)
}

func insert(ctx context.Context, e Enqueuer, tx store.Tx, args river.JobArgs, opts *river.InsertOpts) error {
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, opts)
	}
	return e.Insert(ctx, args, opts)
}
