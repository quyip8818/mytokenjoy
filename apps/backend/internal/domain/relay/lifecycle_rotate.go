package relay

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) SyncRotatePlatformKey(ctx context.Context, platformKeyID string) (string, error) {
	if !l.Enabled() {
		return "", domain.ServiceUnavailable("relay not enabled")
	}
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return "", err
	}
	key, ok := findPlatformKey(platformKeys, platformKeyID)
	if !ok {
		return "", fmt.Errorf("platform key not found")
	}
	if key.Status != "active" {
		return "", domain.Conflict("platform key is not active")
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil || mapping.SyncStatus != store.RelaySyncStatusSynced {
		return "", fmt.Errorf("relay mapping missing for %s", platformKeyID)
	}
	token, err := l.client.RegenerateToken(ctx, *mapping.NewAPITokenID)
	if err != nil {
		return "", err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == platformKeyID {
			platformKeys[i].FullKey = &token.Key
			platformKeys[i].KeyPrefix = relayPlatformKeyPrefix(token.Key)
			if err := l.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return "", err
			}
			break
		}
	}
	return token.Key, nil
}
