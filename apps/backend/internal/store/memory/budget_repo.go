package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryBudgetRepo struct{ store *Store }

func (r *memoryBudgetRepo) Tree(ctx context.Context) ([]types.BudgetNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetTree(r.store.companySnapshot(store.CompanyID(ctx)).BudgetTree), nil
}

func (r *memoryBudgetRepo) SetTree(ctx context.Context, tree []types.BudgetNode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.BudgetTree = store.CloneBudgetTree(tree)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryBudgetRepo) AddGroupConsumed(ctx context.Context, groupID string, amountCNY float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for i := range snap.BudgetGroups {
		if snap.BudgetGroups[i].ID == groupID {
			snap.BudgetGroups[i].Consumed += amountCNY
			r.store.setCompanySnapshot(tid, snap)
			return nil
		}
	}
	return nil
}

func (r *memoryBudgetRepo) RollupDepartmentConsumed(ctx context.Context, departmentID string, amountCNY float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	tree := store.CloneBudgetTree(snap.BudgetTree)
	if err := rollupDepartmentConsumedInTree(tree, departmentID, amountCNY); err != nil {
		return err
	}
	snap.BudgetTree = tree
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryBudgetRepo) GetDepartmentBudget(ctx context.Context, departmentID string) (float64, float64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, false, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	node := pkgbudget.FindBudgetNode(r.store.companySnapshot(store.CompanyID(ctx)).BudgetTree, departmentID)
	if node == nil {
		return 0, 0, false, nil
	}
	return node.Budget, node.Consumed, true, nil
}

func (r *memoryBudgetRepo) GetGroupBudget(ctx context.Context, groupID string) (float64, float64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, false, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, group := range r.store.companySnapshot(store.CompanyID(ctx)).BudgetGroups {
		if group.ID == groupID {
			return group.Budget, group.Consumed, true, nil
		}
	}
	return 0, 0, false, nil
}

func rollupDepartmentConsumedInTree(tree []types.BudgetNode, leafDepartmentID string, costCNY float64) error {
	node := pkgbudget.FindBudgetNode(tree, leafDepartmentID)
	if node == nil {
		return nil
	}
	node.Consumed += costCNY
	for _, ancestorID := range collectAncestorIDsForRollup(tree, leafDepartmentID) {
		ancestor := pkgbudget.FindBudgetNode(tree, ancestorID)
		if ancestor != nil {
			ancestor.Consumed += costCNY
		}
	}
	return nil
}

func collectAncestorIDsForRollup(tree []types.BudgetNode, leafID string) []string {
	var ancestors []string
	var walk func(nodes []types.BudgetNode, path []string) bool
	walk = func(nodes []types.BudgetNode, path []string) bool {
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

func (r *memoryBudgetRepo) Groups(ctx context.Context) ([]types.BudgetGroup, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetGroups(r.store.companySnapshot(store.CompanyID(ctx)).BudgetGroups), nil
}

func (r *memoryBudgetRepo) SetGroups(ctx context.Context, groups []types.BudgetGroup) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.BudgetGroups = store.CloneBudgetGroups(groups)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryBudgetRepo) OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	if err := ctx.Err(); err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.companySnapshot(store.CompanyID(ctx)).OverrunPolicy, nil
}

func (r *memoryBudgetRepo) SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.OverrunPolicy = policy
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryBudgetRepo) AlertRules(ctx context.Context) ([]types.AlertRule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneAlertRules(r.store.companySnapshot(store.CompanyID(ctx)).AlertRules), nil
}

func (r *memoryBudgetRepo) SetAlertRules(ctx context.Context, rules []types.AlertRule) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.AlertRules = store.CloneAlertRules(rules)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}
