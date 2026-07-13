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
	if err := l.persistPlatformKeySecret(ctx, platformKeyID, token.Key); err != nil {
		return "", err
	}
	return token.Key, nil
}
