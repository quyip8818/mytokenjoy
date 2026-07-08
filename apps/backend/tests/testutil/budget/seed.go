package budgetfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func SeedDeptOverrun(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	testutil.SetDeptSnapshotConsumed(t, st, deptID, consumed)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
}
