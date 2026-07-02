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

func (r *memoryKeysRepo) AddPlatformKeyUsed(ctx context.Context, keyID string, amountCNY float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for i := range snap.PlatformKeys {
		if snap.PlatformKeys[i].ID == keyID {
			snap.PlatformKeys[i].Used += amountCNY
			r.store.setCompanySnapshot(tid, snap)
			return nil
		}
	}
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

func (r *memoryKeysRepo) PlatformKeyByID(ctx context.Context, keyID string) (*types.PlatformKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, key := range r.store.companySnapshot(store.CompanyID(ctx)).PlatformKeys {
		if key.ID == keyID {
			cloned := store.ClonePlatformKey(key)
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *memoryKeysRepo) SumMemberKeyUsed(ctx context.Context, memberID string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var total float64
	for _, key := range r.store.companySnapshot(store.CompanyID(ctx)).PlatformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil {
			total += key.Used
		}
	}
	return total, nil
}

func (r *memoryKeysRepo) ListActiveMemberKeys(ctx context.Context, memberID string) ([]types.PlatformKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]types.PlatformKey, 0)
	for _, key := range r.store.companySnapshot(store.CompanyID(ctx)).PlatformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			out = append(out, store.ClonePlatformKey(key))
		}
	}
	return out, nil
}
