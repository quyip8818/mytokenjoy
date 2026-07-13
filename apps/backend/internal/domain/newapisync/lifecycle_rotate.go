package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
)

func (l *NewAPISync) SyncRotatePlatformKey(ctx context.Context, platformKeyID string) (string, error) {
	if !l.Enabled() {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	_, mapping, err := l.requireSyncedMapping(ctx, platformKeyID)
	if err != nil {
		return "", err
	}
	token, err := l.client.RegenerateToken(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return "", err
	}
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return "", err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == platformKeyID {
			platformKeys[i].FullKey = &token.Key
			platformKeys[i].KeyPrefix = newAPIPlatformKeyPrefix(token.Key)
			if err := l.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return "", err
			}
			break
		}
	}
	return token.Key, nil
}
