package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryModelsRepo struct{ store *Store }

func (r *memoryModelsRepo) Models(ctx context.Context) ([]types.ModelInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneModels(r.store.companySnapshot(store.CompanyID(ctx)).Models), nil
}

func (r *memoryModelsRepo) SetModels(ctx context.Context, models []types.ModelInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Models = store.CloneModels(models)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryModelsRepo) RoutingRules(ctx context.Context) ([]types.RoutingRule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoutingRules(r.store.companySnapshot(store.CompanyID(ctx)).RoutingRules), nil
}

func (r *memoryModelsRepo) SetRoutingRules(ctx context.Context, rules []types.RoutingRule) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.RoutingRules = store.CloneRoutingRules(rules)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}
