package memory

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgRepo struct{ store *Store }

func (r *memoryOrgRepo) DataSourceStatus(ctx context.Context) (types.DataSourceStatus, error) {
	if err := ctx.Err(); err != nil {
		return types.DataSourceStatus{}, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.companySnapshot(store.CompanyID(ctx)).DataSourceStatus, nil
}

func (r *memoryOrgRepo) SetDataSourceStatus(ctx context.Context, status types.DataSourceStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.DataSourceStatus = status
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) ImportFailures(ctx context.Context) ([]types.ImportFailure, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneImportFailures(r.store.companySnapshot(store.CompanyID(ctx)).ImportFailures), nil
}

func (r *memoryOrgRepo) SetImportFailures(ctx context.Context, failures []types.ImportFailure) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.ImportFailures = store.CloneImportFailures(failures)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) SyncConfig(ctx context.Context) (types.SyncConfig, error) {
	if err := ctx.Err(); err != nil {
		return types.SyncConfig{}, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.companySnapshot(store.CompanyID(ctx)).SyncConfig, nil
}

func (r *memoryOrgRepo) SetSyncConfig(ctx context.Context, cfg types.SyncConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.SyncConfig = cfg
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) SyncLogs(ctx context.Context) ([]types.SyncLog, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneSyncLogs(r.store.companySnapshot(store.CompanyID(ctx)).SyncLogs), nil
}

func (r *memoryOrgRepo) AppendSyncLog(ctx context.Context, log types.SyncLog) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.SyncLogs = append([]types.SyncLog{log}, snap.SyncLogs...)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) Departments(ctx context.Context) ([]types.Department, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneDepartments(r.store.companySnapshot(store.CompanyID(ctx)).Departments), nil
}

func (r *memoryOrgRepo) SetDepartments(ctx context.Context, departments []types.Department) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Departments = store.CloneDepartments(departments)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMembers(r.store.companySnapshot(store.CompanyID(ctx)).Members), nil
}

func (r *memoryOrgRepo) SetMembers(ctx context.Context, members []types.Member) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Members = store.CloneMembers(members)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func memberPasswordKey(companyID int64, memberID string) string {
	return fmt.Sprintf("%d:%s", companyID, memberID)
}

func (r *memoryOrgRepo) SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.memberPasswordHashes == nil {
		r.store.memberPasswordHashes = make(map[string]string)
	}
	r.store.memberPasswordHashes[memberPasswordKey(store.CompanyID(ctx), memberID)] = passwordHash
	return nil
}

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
