package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryBudgetRepo struct{ store *Store }

func (r *memoryBudgetRepo) Tree() []types.BudgetNode {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetTree(r.store.data.BudgetTree)
}

func (r *memoryBudgetRepo) SetTree(tree []types.BudgetNode) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetTree = store.CloneBudgetTree(tree)
	return nil
}

func (r *memoryBudgetRepo) Groups() []types.BudgetGroup {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetGroups(r.store.data.BudgetGroups)
}

func (r *memoryBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetGroups = store.CloneBudgetGroups(groups)
	return nil
}

func (r *memoryBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.OverrunPolicy
}

func (r *memoryBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.OverrunPolicy = policy
	return nil
}

func (r *memoryBudgetRepo) AlertRules() []types.AlertRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneAlertRules(r.store.data.AlertRules)
}

func (r *memoryBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AlertRules = store.CloneAlertRules(rules)
	return nil
}

func (r *memoryBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMemberQuotaPools(r.store.data.MemberQuotaPools)
}

func (r *memoryBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.MemberQuotaPools = store.CloneMemberQuotaPools(pools)
	return nil
}
