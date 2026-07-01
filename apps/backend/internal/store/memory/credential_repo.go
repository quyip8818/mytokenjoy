package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryCredentialRepo struct {
	store *Store
}

func (r *memoryCredentialRepo) GetCredential(ctx context.Context) (*types.StoredCredential, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	snap := r.store.companySnapshot(store.CompanyID(ctx))
	if snap.CredentialPlatform == nil || len(snap.EncryptedCredential) == 0 {
		return nil, nil
	}
	platform := *snap.CredentialPlatform
	encrypted := make([]byte, len(snap.EncryptedCredential))
	copy(encrypted, snap.EncryptedCredential)
	return &types.StoredCredential{Platform: platform, Encrypted: encrypted}, nil
}

func (r *memoryCredentialRepo) SaveCredential(ctx context.Context, platform types.Platform, encrypted []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	p := platform
	snap.CredentialPlatform = &p
	snap.EncryptedCredential = make([]byte, len(encrypted))
	copy(snap.EncryptedCredential, encrypted)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryCredentialRepo) ClearCredential(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.CredentialPlatform = nil
	snap.EncryptedCredential = nil
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

var _ store.CredentialRepository = (*memoryCredentialRepo)(nil)
