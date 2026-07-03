package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoles(r.store.companySnapshot(store.CompanyID(ctx)).Roles), nil
}

func (r *memoryOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Roles = store.CloneRoles(roles)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) Permissions(ctx context.Context) ([]types.Permission, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePermissions(r.store.permissions), nil
}
