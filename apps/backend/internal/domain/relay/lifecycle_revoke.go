package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("relay not enabled")
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return nil
	}
	if err := l.client.DeleteToken(ctx, *mapping.NewAPITokenID); err != nil {
		return err
	}
	mapping.SyncStatus = store.RelaySyncStatusSynced
	return l.mappings.UpsertMapping(ctx, *mapping)
}
