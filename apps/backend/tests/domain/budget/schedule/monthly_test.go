package schedule_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/budget/schedule"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

type monthTestEnqueuer struct {
	rebalance int
}

func (r *monthTestEnqueuer) InsertRebalance(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	r.rebalance++
	return nil
}

func TestEnsureMonthRebalanceEnqueuesCompanyAxis(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	enqueuer := &monthTestEnqueuer{}
	ctx := testutil.Ctx()

	if err := schedule.EnsureMonthRebalance(ctx, cfg, st, enqueuer, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if enqueuer.rebalance != 1 {
		t.Fatalf("expected 1 rebalance enqueue, got %d", enqueuer.rebalance)
	}

	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if err := st.TenantBackgroundState().SetLastRebalancedPeriod(ctx, contract.DefaultCompanyID, current); err != nil {
		t.Fatal(err)
	}
	enqueuer.rebalance = 0
	if err := schedule.EnsureMonthRebalance(ctx, cfg, st, enqueuer, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if enqueuer.rebalance != 0 {
		t.Fatalf("expected skip when period current, got %d", enqueuer.rebalance)
	}
}

func TestTenantWatchdogKindRegistered(t *testing.T) {
	t.Parallel()
	var args jobs.TenantWatchdogArgs
	if args.Kind() != jobs.KindTenantWatchdog {
		t.Fatalf("expected kind %q, got %q", jobs.KindTenantWatchdog, args.Kind())
	}
}
