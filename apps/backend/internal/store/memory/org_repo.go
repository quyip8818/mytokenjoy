package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgRepo struct{ store *Store }

func (r *memoryOrgRepo) DataSourceStatus() types.DataSourceStatus {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.DataSourceStatus
}

func (r *memoryOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.DataSourceStatus = status
	return nil
}

func (r *memoryOrgRepo) ImportFailures() []types.ImportFailure {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneImportFailures(r.store.data.ImportFailures)
}

func (r *memoryOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ImportFailures = store.CloneImportFailures(failures)
	return nil
}

func (r *memoryOrgRepo) SyncConfig() types.SyncConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.SyncConfig
}

func (r *memoryOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.SyncConfig = cfg
	return nil
}

func (r *memoryOrgRepo) SyncLogs() []types.SyncLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneSyncLogs(r.store.data.SyncLogs)
}

func (r *memoryOrgRepo) AppendSyncLog(log types.SyncLog) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.SyncLogs = append([]types.SyncLog{log}, r.store.data.SyncLogs...)
	return nil
}

func (r *memoryOrgRepo) Departments() []types.Department {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneDepartments(r.store.data.Departments)
}

func (r *memoryOrgRepo) SetDepartments(departments []types.Department) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Departments = store.CloneDepartments(departments)
	return nil
}

func (r *memoryOrgRepo) Members() []types.Member {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMembers(r.store.data.Members)
}

func (r *memoryOrgRepo) SetMembers(members []types.Member) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Members = store.CloneMembers(members)
	return nil
}

func (r *memoryOrgRepo) Roles() []types.Role {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoles(r.store.data.Roles)
}

func (r *memoryOrgRepo) SetRoles(roles []types.Role) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Roles = store.CloneRoles(roles)
	return nil
}

func (r *memoryOrgRepo) Permissions() []types.Permission {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePermissions(r.store.data.Permissions)
}
