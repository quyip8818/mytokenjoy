package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPIKeyID == nil {
		return nil
	}
	if err := l.client.DeleteToken(ctx, *mapping.NewAPIKeyID); err != nil {
		return err
	}
	mapping.SyncStatus = store.MappingSyncStatusSynced
	return l.mappings.UpsertMapping(ctx, *mapping)
}
