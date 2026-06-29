package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

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

var _ store.CredentialRepository = (*memoryCredentialRepo)(nil)
