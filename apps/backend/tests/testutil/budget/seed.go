//go:build testhook

package budgetfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func SeedDeptOverrun(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	testutil.SetDeptSnapshotConsumed(t, st, deptID, consumed)
	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
}
