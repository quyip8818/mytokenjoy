package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestRebalanceRefreshesCombinedKeyRemain(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	rebalance := budget.NewRebalanceService(cfg, st)
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

	// Rebalance should succeed — it refreshes local combined_key_remain without remote calls.
	if err := rebalance.ProcessAxis(ctx, store.RebalanceAxisMember, contract.IDMember1); err != nil {
		t.Fatal(err)
	}
}
