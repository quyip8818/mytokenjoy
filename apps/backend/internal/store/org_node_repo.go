package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

// OrgNodeRepository reads and writes org_nodes. department_id columns elsewhere
// reference OrgNode.ID (org_node_id semantics).
type OrgNodeRepository interface {
	Tree(ctx context.Context) ([]types.OrgNode, error)
	SetTree(ctx context.Context, tree []types.OrgNode) error
	GetNodeBudget(ctx context.Context, nodeID uuid.UUID) (budget float64, found bool, err error)
	GetNodePeriod(ctx context.Context, nodeID uuid.UUID) (period string, found bool, err error)
	ListSelfAndAncestorIDs(ctx context.Context, leafNodeID uuid.UUID) ([]uuid.UUID, error)
}

// OrgNodePathLabel returns an ltree-safe label for a node ID (hyphens become underscores).
func OrgNodePathLabel(nodeID uuid.UUID) string {
	return strings.ReplaceAll(nodeID.String(), "-", "_")
}

// ComputeOrgNodePaths returns ltree paths for flattened org nodes.
func ComputeOrgNodePaths(nodes []types.OrgNode) map[uuid.UUID]string {
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	paths := make(map[uuid.UUID]string, len(flat))
	for _, node := range flat {
		label := OrgNodePathLabel(node.ID)
		if node.ParentID == nil || *node.ParentID == uuid.Nil {
			paths[node.ID] = label
			continue
		}
		parentPath, ok := paths[*node.ParentID]
		if !ok {
			paths[node.ID] = label
			continue
		}
		paths[node.ID] = parentPath + "." + label
	}
	return paths
}
