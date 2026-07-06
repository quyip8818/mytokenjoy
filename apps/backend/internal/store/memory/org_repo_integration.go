package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryOrgRepo) Integration(ctx context.Context) (types.OrgIntegration, error) {
	if err := ctx.Err(); err != nil {
		return types.OrgIntegration{}, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneOrgIntegration(r.store.companySnapshot(store.CompanyID(ctx)).OrgIntegration), nil
}

func (r *memoryOrgRepo) SetIntegration(ctx context.Context, integration types.OrgIntegration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	current := snap.OrgIntegration
	current.ApplyDataSourceStatus(integration.ToDataSourceStatus())
	current.ApplySyncConfig(integration.ToSyncConfig())
	if len(integration.FieldMappings) > 0 {
		current.FieldMappings = append([]types.FieldMapping{}, integration.FieldMappings...)
	}
	snap.OrgIntegration = current
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) GetIntegrationCredential(ctx context.Context) (*types.StoredCredential, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.companySnapshot(store.CompanyID(ctx)).OrgIntegration.ToStoredCredential(), nil
}

func (r *memoryOrgRepo) SaveIntegrationCredential(ctx context.Context, platform types.Platform, encrypted []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	p := platform
	snap.OrgIntegration.Platform = &p
	snap.OrgIntegration.EncryptedCredential = make([]byte, len(encrypted))
	copy(snap.OrgIntegration.EncryptedCredential, encrypted)
	snap.OrgIntegration.FieldMappings = nil
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) ClearIntegrationCredential(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.OrgIntegration.EncryptedCredential = nil
	snap.OrgIntegration.FieldMappings = nil
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) FieldMappings(ctx context.Context) ([]types.FieldMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	mappings := r.store.companySnapshot(store.CompanyID(ctx)).OrgIntegration.FieldMappings
	return append([]types.FieldMapping{}, mappings...), nil
}

func (r *memoryOrgRepo) SetFieldMappings(ctx context.Context, mappings []types.FieldMapping) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.OrgIntegration.FieldMappings = append([]types.FieldMapping{}, mappings...)
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
