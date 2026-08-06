package budget

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// enrichTreeConsumed fills BudgetNode.Consumed from ledger aggregation.
// Each leaf department gets its direct consumed; parent nodes accumulate children.
func enrichTreeConsumed(ctx context.Context, ledger store.LedgerRepository, tree []types.BudgetNode, periodKey string) error {
	if len(tree) == 0 || periodKey == "" {
		return nil
	}
	consumed, err := ledger.SumCostAllDepartments(ctx, periodKey)
	if err != nil {
		return err
	}
	applyConsumedToTree(tree, consumed)
	return nil
}

// applyConsumedToTree sets direct consumed on each node and bubbles up to parents.
// Returns the subtree total so parents can accumulate.
func applyConsumedToTree(nodes []types.BudgetNode, consumed map[uuid.UUID]float64) float64 {
	var total float64
	for i := range nodes {
		childTotal := applyConsumedToTree(nodes[i].Children, consumed)
		direct := consumed[nodes[i].ID]
		nodes[i].Consumed = direct + childTotal
		total += nodes[i].Consumed
	}
	return total
}
