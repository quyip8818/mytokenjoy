package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgNodeRepo struct{ store *Store }

func (r *memoryOrgNodeRepo) Tree(ctx context.Context) ([]types.OrgNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneOrgNodes(r.store.companySnapshot(store.CompanyID(ctx)).OrgNodes), nil
}

func (r *memoryOrgNodeRepo) SetTree(ctx context.Context, tree []types.OrgNode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.OrgNodes = store.CloneOrgNodes(tree)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgNodeRepo) RollupConsumed(ctx context.Context, nodeID string, amountCNY float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	tree := store.CloneOrgNodes(snap.OrgNodes)
	if err := rollupOrgNodeConsumedInTree(tree, nodeID, amountCNY); err != nil {
		return err
	}
	snap.OrgNodes = tree
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgNodeRepo) GetNodeBudget(ctx context.Context, nodeID string) (float64, float64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, false, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	node := pkgorg.FindOrgNode(r.store.companySnapshot(store.CompanyID(ctx)).OrgNodes, nodeID)
	if node == nil {
		return 0, 0, false, nil
	}
	return node.Budget, node.Consumed, true, nil
}

func rollupOrgNodeConsumedInTree(tree []types.OrgNode, leafNodeID string, costCNY float64) error {
	node := pkgorg.FindOrgNode(tree, leafNodeID)
	if node == nil {
		return nil
	}
	node.Consumed += costCNY
	for _, ancestorID := range collectOrgNodeAncestorIDs(tree, leafNodeID) {
		ancestor := pkgorg.FindOrgNode(tree, ancestorID)
		if ancestor != nil {
			ancestor.Consumed += costCNY
		}
	}
	return nil
}

func collectOrgNodeAncestorIDs(tree []types.OrgNode, leafID string) []string {
	var ancestors []string
	var walk func(nodes []types.OrgNode, path []string) bool
	walk = func(nodes []types.OrgNode, path []string) bool {
		for _, node := range nodes {
			nextPath := append(path, node.ID)
			if node.ID == leafID {
				ancestors = append([]string{}, path...)
				return true
			}
			if len(node.Children) > 0 && walk(node.Children, nextPath) {
				return true
			}
		}
		return false
	}
	walk(tree, nil)
	return ancestors
}
