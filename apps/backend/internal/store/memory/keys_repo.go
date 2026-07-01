package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryKeysRepo struct{ store *Store }

func (r *memoryKeysRepo) ProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneProviderKeys(r.store.providerKeys), nil
}

func (r *memoryKeysRepo) SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.providerKeys = store.CloneProviderKeys(keys)
	return nil
}

func (r *memoryKeysRepo) PlatformKeys(ctx context.Context) ([]types.PlatformKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePlatformKeys(r.store.companySnapshot(store.CompanyID(ctx)).PlatformKeys), nil
}

func (r *memoryKeysRepo) SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.PlatformKeys = store.ClonePlatformKeys(keys)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryKeysRepo) Approvals(ctx context.Context) ([]types.KeyApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneApprovals(r.store.companySnapshot(store.CompanyID(ctx)).Approvals), nil
}

func (r *memoryKeysRepo) SetApprovals(ctx context.Context, approvals []types.KeyApproval) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Approvals = store.CloneApprovals(approvals)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}
