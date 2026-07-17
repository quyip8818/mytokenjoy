package provision

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/store"
)

// UnreadyPlatformKeyIDs lists active platform keys that are not fully synced to NewAPI.
func UnreadyPlatformKeyIDs(ctx context.Context, d syncdeps.Deps) ([]uuid.UUID, error) {
	if !syncdeps.Enabled(d) {
		return nil, nil
	}
	platformKeys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return nil, err
	}
	var unready []uuid.UUID
	for _, key := range platformKeys {
		if key.Status != "active" {
			continue
		}
		mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, key.ID)
		if err != nil {
			return nil, err
		}
		if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
			unready = append(unready, key.ID)
			continue
		}
		hash, ok, err := d.Store.Keys().PlatformKeyHashByID(ctx, key.ID)
		if err != nil {
			return nil, err
		}
		if !ok || hash == store.HashPlatformKey("pending:"+key.ID.String()) {
			unready = append(unready, key.ID)
		}
	}
	return unready, nil
}
