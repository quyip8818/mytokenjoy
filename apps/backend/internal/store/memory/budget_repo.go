package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
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

func (r *memoryBudgetRepo) MemberQuotaPools(ctx context.Context) (map[string]types.MemberQuotaPool, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMemberQuotaPools(r.store.companySnapshot(store.CompanyID(ctx)).MemberQuotaPools), nil
}

func (r *memoryBudgetRepo) SetMemberQuotaPools(ctx context.Context, pools map[string]types.MemberQuotaPool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.MemberQuotaPools = store.CloneMemberQuotaPools(pools)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}
