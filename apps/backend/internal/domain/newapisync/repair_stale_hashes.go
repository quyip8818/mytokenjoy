package newapisync

import (
	"context"
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/store"
)

// repairStalePlatformKeyHashes fixes demo keys synced before key_hash was persisted.
func (l *NewAPISync) repairStalePlatformKeyHashes(ctx context.Context) error {
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for _, key := range platformKeys {
		if key.Status != "active" {
			continue
		}
		hash, ok, err := l.store.Keys().PlatformKeyHashByID(ctx, key.ID)
		if err != nil {
			return err
		}
		if !ok || hash != store.HashPlatformKey("pending:"+key.ID) {
			continue
		}
		mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, key.ID)
		if err != nil {
			return err
		}
		if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
			continue
		}
		bearer, err := l.client.GetTokenKey(ctx, *mapping.NewAPIKeyID)
		if err != nil {
			return fmt.Errorf("fetch token key for %s: %w", key.ID, err)
		}
		if bearer == "" || strings.Contains(bearer, "*") {
			continue
		}
		if err := l.persistPlatformKeySecret(ctx, key.ID, bearer); err != nil {
			return err
		}
	}
	return nil
}
