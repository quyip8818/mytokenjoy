package orgfix

import (
	"github.com/tokenjoy/backend/internal/domain/types"
)

// MutateOrgNode recursively traverses the OrgNode tree and applies fn to the
// node matching the given id. Returns true if the node was found and mutated.
func MutateOrgNode(nodes []types.OrgNode, id string, fn func(*types.OrgNode)) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			fn(&nodes[i])
			return true
		}
		if len(nodes[i].Children) > 0 && MutateOrgNode(nodes[i].Children, id, fn) {
			return true
		}
	}
	return false
}

// SetNodePeriod sets the Period field of the OrgNode matching id.
func SetNodePeriod(nodes []types.OrgNode, id string, period string) bool {
	return MutateOrgNode(nodes, id, func(n *types.OrgNode) {
		n.Period = period
	})
}

// SetNodeBudget sets the Budget field of the OrgNode matching id.
func SetNodeBudget(nodes []types.OrgNode, id string, budget float64) bool {
	return MutateOrgNode(nodes, id, func(n *types.OrgNode) {
		n.Budget = budget
	})
}
