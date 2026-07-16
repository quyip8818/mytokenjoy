package memberanalytics_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/adapter"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/runtime"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newMemberAnalyticsService(t *testing.T) (domainmemberanalytics.Service, context.Context) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.CtxForCompany(contract.DefaultCompanyID)
	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	newAPISync := newapisync.New(cfg, st, nil, nil, policy.NewChannelPolicy(cfg), adapter.NewNewAPISyncEnqueuer(jobs.NoopEnqueuer{}))
	keysSvc := domainkeys.NewService(cfg, st, newAPISync, common.NewDelayer(false))
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	return domainmemberanalytics.NewService(cfg, keysSvc, reader), ctx
}

func TestGetDashboardReturnsUsageForMember(t *testing.T) {
	t.Parallel()
	svc, ctx := newMemberAnalyticsService(t)
	view, err := svc.GetDashboard(ctx, contract.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if view.UsageStats.RequestCount <= 0 {
		t.Fatalf("expected positive request count, got %+v", view.UsageStats)
	}
	if view.ResourceConsumption.TotalCost <= 0 {
		t.Fatalf("expected positive total cost, got %+v", view.ResourceConsumption)
	}
}
