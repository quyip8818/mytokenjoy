package memberanalytics_test

import (
	"context"
	"testing"

	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
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
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil, nil, relay.NewChannelPolicy(cfg))
	keysSvc := domainkeys.NewService(cfg, st, lifecycle, common.NewDelayer(false))
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
