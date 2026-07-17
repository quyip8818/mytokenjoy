//go:build testhook

package budgetfix

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// EnsureMonthRebalanceCurrent marks the tenant as rebalanced for the cfg clock month so
// budget.Projector batch tests are not blocked by EnsureMonthRebalance side effects.
func EnsureMonthRebalanceCurrent(t *testing.T, ctx context.Context, cfg config.Config, st store.Store, companyID uuid.UUID) {
	t.Helper()
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if err := st.TenantBackgroundState().EnsureRow(ctx, companyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().SetLastRebalancedPeriod(ctx, companyID, current); err != nil {
		t.Fatal(err)
	}
}
