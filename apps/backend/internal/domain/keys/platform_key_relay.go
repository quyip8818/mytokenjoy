package keys

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) requireRelay() error {
	if s.relaySync == nil || !s.relaySync.Enabled() {
		return domain.ServiceUnavailable("Relay is required for platform keys")
	}
	return nil
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
	if err := s.relaySync.SyncCreatePlatformKey(ctx, created, departmentID); err != nil {
		return types.PlatformKey{}, fmt.Errorf("relay sync enqueue: %w", err)
	}
	fullKey, err := s.relaySync.TrySyncCreate(ctx, created.ID)
	if err != nil {
		s.relaySync.RollbackFailedCreate(ctx, created.ID)
		return types.PlatformKey{}, domain.ServiceUnavailable("Relay sync failed")
	}
	if err := s.updatePlatformKeyFullKey(ctx, created.ID, fullKey); err != nil {
		return types.PlatformKey{}, err
	}
	refreshed, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	for _, key := range refreshed {
		if key.ID == created.ID {
			return s.enrichPlatformKeyResponse(ctx, key)
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
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
	return "", domain.Validation("department required for relay sync")
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
