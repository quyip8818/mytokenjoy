package platformkey

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func findPlatformKey(keys []types.PlatformKey, id uuid.UUID) (types.PlatformKey, bool) {
	for _, key := range keys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}

func RequireSyncedMapping(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) (types.PlatformKey, store.PlatformKeyMapping, error) {
	platformKeys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, err
	}
	key, ok := findPlatformKey(platformKeys, platformKeyID)
	if !ok {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, fmt.Errorf("platform key not found")
	}
	if key.Status != "active" {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, domain.Conflict("platform key is not active")
	}
	mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, err
	}
	if mapping == nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, domain.ServiceUnavailable("platform key not synced to NewAPI yet")
	}
	if mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, domain.ServiceUnavailable("platform key not synced to NewAPI yet")
	}
	return key, *mapping, nil
}
