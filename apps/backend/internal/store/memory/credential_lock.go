package memory

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memorySchedulerLockRepo struct {
	store *Store
}

type schedulerLockEntry struct {
	holder     string
	leaseUntil time.Time
}

func (r *memorySchedulerLockRepo) TryAcquire(
	ctx context.Context,
	lockName, holder string,
	lease time.Duration,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.schedulerLocks == nil {
		r.store.schedulerLocks = make(map[string]schedulerLockEntry)
	}
	now := time.Now()
	entry, ok := r.store.schedulerLocks[lockName]
	if ok && entry.leaseUntil.After(now) && entry.holder != holder {
		return false, nil
	}
	r.store.schedulerLocks[lockName] = schedulerLockEntry{
		holder:     holder,
		leaseUntil: now.Add(lease),
	}
	return true, nil
}

func (r *memorySchedulerLockRepo) Release(ctx context.Context, lockName, holder string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	entry, ok := r.store.schedulerLocks[lockName]
	if !ok || entry.holder != holder {
		return nil
	}
	delete(r.store.schedulerLocks, lockName)
	return nil
}

type memoryCredentialRepo struct {
	store *Store
}

func (r *memoryCredentialRepo) GetCredential() (*types.StoredCredential, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	if r.store.data.CredentialPlatform == nil || len(r.store.data.EncryptedCredential) == 0 {
		return nil, nil
	}
	platform := *r.store.data.CredentialPlatform
	encrypted := make([]byte, len(r.store.data.EncryptedCredential))
	copy(encrypted, r.store.data.EncryptedCredential)
	return &types.StoredCredential{Platform: platform, Encrypted: encrypted}, nil
}

func (r *memoryCredentialRepo) SaveCredential(platform types.Platform, encrypted []byte) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	p := platform
	r.store.data.CredentialPlatform = &p
	r.store.data.EncryptedCredential = make([]byte, len(encrypted))
	copy(r.store.data.EncryptedCredential, encrypted)
	return nil
}

func (r *memoryCredentialRepo) ClearCredential() error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.CredentialPlatform = nil
	r.store.data.EncryptedCredential = nil
	return nil
}

var _ store.SchedulerLockRepository = (*memorySchedulerLockRepo)(nil)
var _ store.CredentialRepository = (*memoryCredentialRepo)(nil)
