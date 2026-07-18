package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestRebalanceBidirectional(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	rebalance := budget.NewRebalanceService(cfg, st, newapi.NewAdminPortAdapter(stub))
	ctx := testutil.Ctx()

	tokenID := int64(42)
	memberID := contract.IDMember1
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID:     contract.DefaultCompanyID,
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		MemberID:      &memberID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	if err := rebalance.ProcessAxis(ctx, store.RebalanceAxisMember, contract.IDMember1); err != nil {
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
		if keys[i].ID == contract.IDPlatformKey1 {
			keys[i].Budget = 1000
			keys[i].Consumed = 999
			if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
				t.Fatal(err)
			}
			break
		}
	}

	if err := rebalance.ProcessAxis(ctx, store.RebalanceAxisMember, contract.IDMember1); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 1 {
		t.Fatalf("expected one update when remain would decrease, calls=%d", stub.UpdateTokenCalls)
	}
}
