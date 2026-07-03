package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryModelsRepo struct {
	store     *Store
	allowlist *memoryModelAllowlistRepo
}

func (r *memoryModelsRepo) Allowlist() store.ModelAllowlistRepository {
	return r.allowlist
}

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

func (r *memoryModelsRepo) ModelByName(ctx context.Context, name string) (*types.ModelInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, model := range r.store.companySnapshot(store.CompanyID(ctx)).Models {
		if model.Name == name {
			cloned := store.CloneModels([]types.ModelInfo{model})[0]
			return &cloned, nil
		}
	}
	return nil, nil
}
