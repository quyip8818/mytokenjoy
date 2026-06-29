package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryKeysRepo struct{ store *Store }

func (r *memoryKeysRepo) ProviderKeys() []types.ProviderKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneProviderKeys(r.store.data.ProviderKeys)
}

func (r *memoryKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ProviderKeys = store.CloneProviderKeys(keys)
	return nil
}

func (r *memoryKeysRepo) PlatformKeys() []types.PlatformKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePlatformKeys(r.store.data.PlatformKeys)
}

func (r *memoryKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.PlatformKeys = store.ClonePlatformKeys(keys)
	return nil
}

func (r *memoryKeysRepo) Approvals() []types.KeyApproval {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneApprovals(r.store.data.Approvals)
}

func (r *memoryKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Approvals = store.CloneApprovals(approvals)
	return nil
}
