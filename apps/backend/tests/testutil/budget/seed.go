package budgetfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func SeedDeptOverrun(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	ctx := testutil.Ctx()
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, deptID, consumed)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
}
