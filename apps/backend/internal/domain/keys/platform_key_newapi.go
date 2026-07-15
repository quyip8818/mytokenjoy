package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
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
	if updated.Budget != previous.Budget {
		if err := domainbudget.RefreshPlatformKeyCombined(ctx, s.store, id, s.cfg.Clock(), nil); err != nil {
			return types.PlatformKey{}, err
		}
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

func (s *service) syncPlatformKeyCreate(ctx context.Context, created types.PlatformKey, departmentID string) (types.PlatformKey, error) {
	fullKey, err := s.newAPISync.SyncPlatformKeyCreate(ctx, created, departmentID)
	if err != nil {
		return types.PlatformKey{}, domain.ServiceUnavailable("NewAPI sync failed")
	}
	created.FullKey = &fullKey
	created.KeyPrefix = platformKeyPrefix(fullKey)
	return s.enrichPlatformKeyResponse(ctx, created)
}

func (s *service) resolvePlatformKeyDepartmentID(
	input types.CreatePlatformKeyInput,
	members []types.Member,
	projects []types.Project,
) (string, error) {
	if input.MemberID != nil {
		if member, ok := org.FindMemberByID(members, *input.MemberID); ok && member.DepartmentID != "" {
			return member.DepartmentID, nil
		}
	}
	if input.ProjectID != nil {
		for _, project := range projects {
			if project.ID == *input.ProjectID && project.OwnerDepartmentID != "" {
				return project.OwnerDepartmentID, nil
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
