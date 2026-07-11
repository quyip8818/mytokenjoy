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

func InsertOrgSync(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := OrgSyncArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}

func InsertMonthlyRebalance(ctx context.Context, e Enqueuer, tx store.Tx) error {
	args := MonthlyRebalanceArgs{}
	if tx != nil {
		return e.InsertInTx(ctx, tx, args, nil)
	}
	return e.Insert(ctx, args, nil)
}
