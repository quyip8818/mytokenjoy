package memory

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgRepo struct {
	store *Store
	nodes *memoryOrgNodeRepo
}

func (r *memoryOrgRepo) Nodes() store.OrgNodeRepository {
	return r.nodes
}

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

func (r *memoryOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMembers(r.store.companySnapshot(store.CompanyID(ctx)).Members), nil
}

func (r *memoryOrgRepo) MemberByID(ctx context.Context, memberID string) (*types.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, member := range r.store.companySnapshot(store.CompanyID(ctx)).Members {
		if member.ID == memberID {
			cloned := store.CloneMember(member)
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *memoryOrgRepo) MemberPersonalQuota(ctx context.Context, memberID string) (float64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, member := range r.store.companySnapshot(store.CompanyID(ctx)).Members {
		if member.ID == memberID {
			return member.PersonalQuota, true, nil
		}
	}
	return 0, false, nil
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

func (r *memoryOrgRepo) UpdateMemberPersonalQuota(ctx context.Context, memberID string, personalQuota float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for i := range snap.Members {
		if snap.Members[i].ID == memberID {
			snap.Members[i].PersonalQuota = personalQuota
			r.store.setCompanySnapshot(tid, snap)
			return nil
		}
	}
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
