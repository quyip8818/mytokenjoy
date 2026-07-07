package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newIngestStore(t *testing.T) (config.Config, store.Store) {
	t.Helper()
	return testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
}

func findDept3(tree []types.BudgetNode) *types.BudgetNode {
	var walk func([]types.BudgetNode) *types.BudgetNode
	walk = func(nodes []types.BudgetNode) *types.BudgetNode {
		for i := range nodes {
			if nodes[i].ID == seed.IDDept3 {
				return &nodes[i]
			}
			if len(nodes[i].Children) > 0 {
				if found := walk(nodes[i].Children); found != nil {
					return found
				}
			}
		}
		return nil
	}
	return walk(tree)
}
