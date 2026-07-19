package usage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

// cachedOrgNodes serves a pre-loaded org tree for ingest entry build (read-only).
type cachedOrgNodes struct {
	tree []types.OrgNode
}

func (c cachedOrgNodes) Tree(context.Context) ([]types.OrgNode, error) {
	return c.tree, nil
}

func (c cachedOrgNodes) SetTree(context.Context, []types.OrgNode) error {
	return fmt.Errorf("cached org nodes: read-only")
}

func (c cachedOrgNodes) GetNodeBudget(_ context.Context, nodeID uuid.UUID) (int64, bool, error) {
	node := pkgorg.FindOrgNode(c.tree, nodeID)
	if node == nil {
		return 0, false, nil
	}
	return node.Budget, true, nil
}

func (c cachedOrgNodes) GetNodePeriod(_ context.Context, nodeID uuid.UUID) (string, bool, error) {
	node := pkgorg.FindOrgNode(c.tree, nodeID)
	if node == nil {
		return "", false, nil
	}
	return node.Period, true, nil
}

func (c cachedOrgNodes) ListSelfAndAncestorIDs(_ context.Context, leafNodeID uuid.UUID) ([]uuid.UUID, error) {
	if leafNodeID == uuid.Nil {
		return nil, nil
	}
	ids := []uuid.UUID{leafNodeID}
	current := leafNodeID
	for {
		node := pkgorg.FindOrgNode(c.tree, current)
		if node == nil || node.ParentID == nil || *node.ParentID == uuid.Nil {
			break
		}
		current = *node.ParentID
		ids = append(ids, current)
	}
	return ids, nil
}

var _ store.OrgNodeRepository = cachedOrgNodes{}
