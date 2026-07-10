package keys

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) requireNewAPI() error {
	if s.newAPISync == nil || !s.newAPISync.Enabled() {
		return domain.ServiceUnavailable("NewAPI is required for platform keys")
	}
	return nil
}

func platformKeyIndex(keys []types.PlatformKey, id string) (int, bool) {
	for i := range keys {
		if keys[i].ID == id {
			return i, true
		}
	}
	return -1, false
}

func (s *service) newAPIRevokeKey(ctx context.Context, id string) ([]types.PlatformKey, int, error) {
	if err := s.requireNewAPI(); err != nil {
		return nil, -1, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return nil, -1, err
	}
	idx, ok := platformKeyIndex(platformKeys, id)
	if !ok {
		return nil, -1, domain.NotFound("Not found")
	}
	if err := s.newAPISync.SyncRevokePlatformKey(ctx, id); err != nil {
		return nil, -1, err
	}
	return platformKeys, idx, nil
}

func (s *service) persistPlatformKeyWithNewAPISync(
	ctx context.Context,
	platformKeys []types.PlatformKey,
	idx int,
	updated, previous types.PlatformKey,
	id string,
) (types.PlatformKey, error) {
	platformKeys[idx] = updated
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}
	if err := s.newAPISync.SyncUpdatePlatformKey(ctx, id, nil); err != nil {
		platformKeys[idx] = previous
		if rollbackErr := s.store.Keys().SetPlatformKeys(ctx, platformKeys); rollbackErr != nil {
			return types.PlatformKey{}, rollbackErr
		}
		return types.PlatformKey{}, err
	}
	return s.enrichPlatformKeyResponse(ctx, updated)
}

func platformKeyPrefix(fullKey string) string {
	prefix := fullKey
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	return prefix
}

func (s *service) updatePlatformKeyFullKey(ctx context.Context, keyID string, fullKey string) error {
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == keyID {
			platformKeys[i].FullKey = &fullKey
			platformKeys[i].KeyPrefix = platformKeyPrefix(fullKey)
			return s.store.Keys().SetPlatformKeys(ctx, platformKeys)
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) syncPlatformKeyCreate(ctx context.Context, created types.PlatformKey, departmentID string) (types.PlatformKey, error) {
	if err := s.newAPISync.SyncCreatePlatformKey(ctx, created, departmentID); err != nil {
		return types.PlatformKey{}, fmt.Errorf("newapi sync enqueue: %w", err)
	}
	fullKey, err := s.newAPISync.TrySyncCreate(ctx, created.ID)
	if err != nil {
		s.newAPISync.RollbackFailedCreate(ctx, created.ID)
		return types.PlatformKey{}, domain.ServiceUnavailable("NewAPI sync failed")
	}
	if err := s.updatePlatformKeyFullKey(ctx, created.ID, fullKey); err != nil {
		return types.PlatformKey{}, err
	}
	created.FullKey = &fullKey
	created.KeyPrefix = platformKeyPrefix(fullKey)
	return s.enrichPlatformKeyResponse(ctx, created)
}

func (s *service) resolvePlatformKeyDepartmentID(
	input types.CreatePlatformKeyInput,
	members []types.Member,
	groups []types.BudgetGroup,
) (string, error) {
	if input.MemberID != nil {
		if member, ok := org.FindMemberByID(members, *input.MemberID); ok && member.DepartmentID != "" {
			return member.DepartmentID, nil
		}
	}
	if input.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *input.BudgetGroupID && len(group.DepartmentIDs) > 0 {
				return group.DepartmentIDs[0], nil
			}
		}
	}
	return "", domain.Validation("department required for newapi sync")
}

func resolveMemberName(memberID string, members []types.Member) (string, error) {
	if memberID == "" {
		return "", domain.BadRequest("member id is required")
	}
	for _, member := range members {
		if member.ID == memberID {
			return member.Name, nil
		}
	}
	return "", domain.NotFound("member not found")
}
