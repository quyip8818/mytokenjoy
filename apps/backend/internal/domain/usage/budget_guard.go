package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

func enforceBudgetCap(
	ctx context.Context,
	cfg config.Config,
	st store.ConsumptionWriter,
	mapping *store.PlatformKeyMapping,
	amount float64,
	periodKey string,
) error {
	if amount <= 0 {
		return nil
	}
	remain, err := pkgbudget.RemainForMapping(ctx, pkgbudget.MappingStores{
		Snapshots: st.BudgetSnapshots(),
		OrgNodes:  st.Org().Nodes(),
		Org:       st.Org(),
		Budget:    st.Budget(),
		Keys:      st.Keys(),
		Clock:     clock.OrDefault(cfg.Clock()),
	}, mapping, periodKey)
	if err != nil {
		return err
	}
	if remain < amount {
		return domain.Forbidden("budget exceeded")
	}
	return nil
}
