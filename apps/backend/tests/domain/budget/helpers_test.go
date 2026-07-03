package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newBudgetService(t *testing.T) (budget.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	return budget.NewService(cfg, st, common.NewDelayer(false)), st
}
