package jobs

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func InsertWalletSync(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := WalletSyncArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertRebalance(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64, axisKind, axisID string) error {
	args := RebalanceArgs{
		CompanyID: companyID,
		AxisKind:  axisKind,
		AxisID:    axisID,
	}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertOverrun(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64, payload []byte) error {
	args := OverrunArgs{
		CompanyID: companyID,
		Payload:   payload,
	}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertNewAPISync(ctx context.Context, e Enqueuer, tx store.Tx, args NewAPISyncArgs) error {
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertOrgSync(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := OrgSyncArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertOrgSyncFanout(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := OrgSyncArgs{CompanyID: OrgSyncFanoutCompanyID}
	opts := OrgSyncFanoutInsertOpts()
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, opts)
	}
	return e.Insert(ctx, args, opts)
}

func InsertMonthlyRebalance(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := MonthlyRebalanceArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertBudgetProjection(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := BudgetProjectionArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertBudgetReconcile(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := BudgetReconcileArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertBudgetReconcileFanout(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := BudgetReconcileFanoutArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertDashboardProject(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := DashboardProjectArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertDashboardProjectFanout(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := DashboardProjectFanoutArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertDashboardReconcile(ctx context.Context, e Enqueuer, tx store.Tx, companyID int64) error {
	args := DashboardReconcileArgs{CompanyID: companyID}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertDashboardReconcileFanout(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := DashboardReconcileFanoutArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}
