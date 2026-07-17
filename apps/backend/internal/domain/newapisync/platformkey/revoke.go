package platformkey

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/store"
)

func SyncRevokePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) error {
	if !syncdeps.Enabled(d) {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil {
		return err
	}
	if mapping == nil || mapping.NewAPIKeyID == nil {
		return nil
	}
	if err := d.Client.DeleteToken(ctx, *mapping.NewAPIKeyID); err != nil {
		return err
	}
	mapping.SyncStatus = store.MappingSyncStatusSynced
	return d.Mappings.UpsertMapping(ctx, *mapping)
}
