package newapisync

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) requireSyncedMapping(ctx context.Context, platformKeyID string) (types.PlatformKey, store.PlatformKeyMapping, error) {
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
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
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, err
	}
	if mapping == nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, fmt.Errorf("platform key mapping missing for %s", platformKeyID)
	}
	if mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
		return types.PlatformKey{}, store.PlatformKeyMapping{}, domain.ServiceUnavailable("platform key not synced to NewAPI yet")
	}
	return key, *mapping, nil
}
