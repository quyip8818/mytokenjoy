package app

import (
	"context"
	"log/slog"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

// usageBudgetOps adapts domain/budget functions to the usage.BudgetOps port interface.
type usageBudgetOps struct {
	cache          domainbudget.CombinedKeyCache
	alertPublisher domainbudget.AlertPublisher
	logger         *slog.Logger
}

// NewUsageBudgetOps creates a BudgetOps adapter for the usage domain.
func NewUsageBudgetOps(cache domainbudget.CombinedKeyCache, alertPublisher domainbudget.AlertPublisher, logger *slog.Logger) domainusage.BudgetOps {
	if alertPublisher == nil {
		alertPublisher = domainbudget.NoopAlertPublisher
	}
	return &usageBudgetOps{cache: cache, alertPublisher: alertPublisher, logger: logger}
}

func (a *usageBudgetOps) ConsumptionDeltas(ctx context.Context, entry types.UsageLedgerEntry, open pkgbudget.OpenBudgetPeriod) ([]domainusage.ConsumedDelta, error) {
	deltas, err := domainbudget.ConsumptionDeltas(ctx, nil, entry, open)
	if err != nil {
		return nil, err
	}
	out := make([]domainusage.ConsumedDelta, len(deltas))
	for i, d := range deltas {
		out[i] = domainusage.ConsumedDelta{
			Kind:      d.Kind,
			AxisID:    d.AxisID,
			PeriodKey: d.PeriodKey,
			Amount:    d.Amount,
		}
	}
	return out, nil
}

func (a *usageBudgetOps) RefreshCombinedKeySummaries(ctx context.Context, companyID int64, summaries []store.CombinedKeySummary) {
	domainbudget.RefreshCombinedKeySummaries(ctx, a.cache, a.logger, companyID, summaries)
}

func (a *usageBudgetOps) CheckBudgetAlerts(ctx context.Context, st store.Store, companyID int64, touchedDepts map[string]struct{}) {
	domainbudget.CheckBudgetAlerts(ctx, st, companyID, touchedDepts, a.alertPublisher, a.logger)
}

func (a *usageBudgetOps) ComputeGatewaySummaryUpdates(ctx context.Context, st store.Store, keyIDs map[string]struct{}, clk clock.Clock) ([]store.CombinedKeySummaryUpdate, error) {
	return domainbudget.ComputeGatewaySummaryUpdates(ctx, st, keyIDs, clk)
}

var _ domainusage.BudgetOps = (*usageBudgetOps)(nil)
