package schedule

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type RebalanceEnqueuer interface {
	InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error
}

func EnsureMonthRebalance(ctx context.Context, cfg config.Config, st store.Store, enqueuer RebalanceEnqueuer, companyID int64) error {
	entryCtx := company.WithDefaultCompany(ctx, companyID)
	tbs, err := st.TenantBackgroundState().Get(entryCtx, companyID)
	if err != nil {
		return err
	}
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if tbs != nil && tbs.LastRebalancedPeriod == current {
		return nil
	}
	return enqueuer.InsertRebalance(entryCtx, companyID, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
}
