package platformkey

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
)

func SyncRotatePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) (string, error) {
	if !syncdeps.Enabled(d) {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	_, mapping, err := RequireSyncedMapping(ctx, d, platformKeyID)
	if err != nil {
		return "", err
	}
	token, err := d.Client.RegenerateToken(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return "", err
	}
	if err := persistPlatformKeySecret(ctx, d, platformKeyID, token.Key); err != nil {
		return "", err
	}
	return token.Key, nil
}
