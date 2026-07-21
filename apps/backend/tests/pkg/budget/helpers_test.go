package budget_test

import (
	"sync"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

var (
	snapshotOnce sync.Once
	testSnapshot store.Snapshot
)

func cachedSnapshot() store.Snapshot {
	snapshotOnce.Do(func() {
		testSnapshot = seed.Load(testutil.TestConfig())
	})
	return testSnapshot
}
