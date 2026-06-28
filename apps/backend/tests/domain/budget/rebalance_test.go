package budget_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestRebalanceOnlyDown(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub)
	rebalance := budget.NewRebalanceService(cfg, st, stub, lifecycle)

	tokenID := int64(42)
	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      testutil.StrPtr(seed.IDMember1),
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	if err := rebalance.ProcessAxis(context.Background(), store.RebalanceAxisMember, seed.IDMember1); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 0 {
		t.Fatalf("expected no update when remain would increase, calls=%d", stub.UpdateTokenCalls)
	}

	stub.Token.RemainQuota = 50000
	keys := st.Keys().PlatformKeys()
	for i := range keys {
		if keys[i].ID == seed.IDPlatformKey1 {
			keys[i].Quota = 1000
			keys[i].Used = 999
			st.Keys().SetPlatformKeys(keys)
			break
		}
	}

	if err := rebalance.ProcessAxis(context.Background(), store.RebalanceAxisMember, seed.IDMember1); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 1 {
		t.Fatalf("expected one update when remain would decrease, calls=%d", stub.UpdateTokenCalls)
	}
}
