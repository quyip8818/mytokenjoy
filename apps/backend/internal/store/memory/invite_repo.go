package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryInviteRepo struct {
	store *Store
}

func (r *memoryInviteRepo) CreateInvite(ctx context.Context, invite store.CompanyInvite) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.invites == nil {
		r.store.invites = make(map[string]store.CompanyInvite)
	}
	r.store.invites[invite.Token] = invite
	return nil
}

func (r *memoryInviteRepo) GetInviteByToken(ctx context.Context, token string) (*store.CompanyInvite, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	inv, ok := r.store.invites[token]
	if !ok {
		return nil, fmt.Errorf("invite not found")
	}
	copy := inv
	return &copy, nil
}

func (r *memoryInviteRepo) MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for token, inv := range r.store.invites {
		if inv.ID == id {
			inv.AcceptedAt = &acceptedAt
			r.store.invites[token] = inv
			return nil
		}
	}
	return fmt.Errorf("invite not found")
}

var _ store.InviteRepository = (*memoryInviteRepo)(nil)
