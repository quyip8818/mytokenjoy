package member_test

import (
	"context"
	"testing"

	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmember "github.com/tokenjoy/backend/internal/domain/member"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newMemberService(t *testing.T) (domainmember.Service, context.Context) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.CtxForCompany(seed.DefaultCompanyID)
	if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil, nil, relay.NewChannelPolicy(cfg))
	keysSvc := domainkeys.NewService(cfg, st, lifecycle, common.NewDelayer(false))
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	return domainmember.NewService(cfg, keysSvc, reader), ctx
}

func TestGetDashboardReturnsUsageForMember(t *testing.T) {
	svc, ctx := newMemberService(t)
	view, err := svc.GetDashboard(ctx, seed.IDMember1)
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
