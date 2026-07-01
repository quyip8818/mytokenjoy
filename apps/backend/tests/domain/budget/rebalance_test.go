package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestRebalanceBidirectional(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	rebalance := budget.NewRebalanceService(cfg, st, stub, lifecycle)
	ctx := testutil.Ctx()

	tokenID := int64(42)
	remainQuota := int64(1000)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		CompanyID:        seed.DefaultCompanyID,
		PlatformKeyID:    seed.IDPlatformKey1,
		NewAPITokenID:    &tokenID,
		MemberID:         testutil.StrPtr(seed.IDMember1),
		DepartmentID:     seed.IDDept3,
		SyncStatus:       store.RelaySyncStatusSynced,
		RelayGroup:       "dept-dept-3",
		RelayRemainQuota: &remainQuota,
	}); err != nil {
		t.Fatal(err)
	}

	if err := rebalance.ProcessAxis(ctx, store.RebalanceAxisMember, seed.IDMember1); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 1 {
		t.Fatalf("expected update when remain would increase, calls=%d", stub.UpdateTokenCalls)
	}

	stub.UpdateTokenCalls = 0
	stub.Token.RemainQuota = 50000
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == seed.IDPlatformKey1 {
			keys[i].Quota = 1000
			keys[i].Used = 999
			if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
				t.Fatal(err)
			}
			break
		}
	}

	if err := rebalance.ProcessAxis(ctx, store.RebalanceAxisMember, seed.IDMember1); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 1 {
		t.Fatalf("expected one update when remain would decrease, calls=%d", stub.UpdateTokenCalls)
	}
}
