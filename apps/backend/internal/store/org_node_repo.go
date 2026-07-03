package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

// OrgNodeRepository reads and writes org_nodes. department_id columns elsewhere
// reference OrgNode.ID (org_node_id semantics).
type OrgNodeRepository interface {
	Tree(ctx context.Context) ([]types.OrgNode, error)
	SetTree(ctx context.Context, tree []types.OrgNode) error
	RollupConsumed(ctx context.Context, nodeID string, amountCNY float64) error
	GetNodeBudget(ctx context.Context, nodeID string) (budget, consumed float64, found bool, err error)
}
