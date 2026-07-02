package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryRelayRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return nil, nil
	}
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) GetMappingByFullKey(ctx context.Context, fullKey string) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for companyID, snap := range r.store.dataByCompany {
		for _, key := range snap.PlatformKeys {
			if key.FullKey == nil || *key.FullKey != fullKey {
				continue
			}
			r.mu.Lock()
			m, ok := r.data.mappings[key.ID]
			r.mu.Unlock()
			if !ok || m.CompanyID != companyID {
				continue
			}
			copy := m
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRelayRepo) GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.data.tokenIndex[tokenID]
	if !ok {
		return nil, nil
	}
	m := r.data.mappings[id]
	if !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return nil, nil
	}
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.data.tokenIndex[tokenID]
	if !ok {
		return nil, nil
	}
	m := r.data.mappings[id]
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) ListMappingsByMemberID(ctx context.Context, memberID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.MemberID != nil && *m.MemberID == memberID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.DepartmentID == departmentID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.BudgetGroupID != nil && *m.BudgetGroupID == budgetGroupID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListActiveMappings(ctx context.Context) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.SyncStatus == store.RelaySyncStatusSynced {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, companyID) {
			continue
		}
		if m.SyncStatus == store.RelaySyncStatusSynced {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) UpsertMapping(ctx context.Context, mapping store.RelayMapping) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	mapping.CompanyID = store.CompanyID(ctx)
	r.data.mappings[mapping.PlatformKeyID] = mapping
	if mapping.NewAPITokenID != nil {
		r.data.tokenIndex[*mapping.NewAPITokenID] = mapping.PlatformKeyID
	}
	return nil
}

func (r *memoryRelayRepo) UpdateMappingSync(
	ctx context.Context,
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.NewAPITokenID = &tokenID
	m.SyncStatus = status
	m.SyncedAt = &syncedAt
	m.NewAPITokenRemainQuota = remainQuota
	r.data.mappings[platformKeyID] = m
	r.data.tokenIndex[tokenID] = platformKeyID
	return nil
}

func (r *memoryRelayRepo) UpdateMappingNewAPITokenRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.NewAPITokenRemainQuota = &remainQuota
	r.data.mappings[platformKeyID] = m
	return nil
}
