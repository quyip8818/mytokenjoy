package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryBudgetRepo struct{ store *Store }

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

func (r *memoryBudgetRepo) BudgetApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetApprovals(r.store.companySnapshot(store.CompanyID(ctx)).BudgetApprovals), nil
}

func (r *memoryBudgetRepo) SetBudgetApprovals(ctx context.Context, items []types.BudgetApproval) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.BudgetApprovals = store.CloneBudgetApprovals(items)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryBudgetRepo) UpdateBudgetApproval(ctx context.Context, id, status string, rejectReason *string, resolvedAt time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for i := range snap.BudgetApprovals {
		if snap.BudgetApprovals[i].ID != id {
			continue
		}
		snap.BudgetApprovals[i].Status = status
		s := resolvedAt.Format("2006-01-02 15:04")
		snap.BudgetApprovals[i].ResolvedAt = &s
		if status == "rejected" {
			snap.BudgetApprovals[i].RejectReason = rejectReason
		}
		r.store.setCompanySnapshot(tid, snap)
		return nil
	}
	return fmt.Errorf("budget approval not found")
}
