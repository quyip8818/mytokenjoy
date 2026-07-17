package platformkey

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
)

func ResolvePlatformKeyBearer(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) (string, error) {
	if !syncdeps.Enabled(d) {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	_, mapping, err := RequireSyncedMapping(ctx, d, platformKeyID)
	if err != nil {
		return "", err
	}
	bearer, err := d.Client.GetTokenKey(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return "", err
	}
	if bearer == "" || strings.Contains(bearer, "*") {
		return "", domain.Conflict("platform key secret unavailable; rotate once via /keys/platform")
	}
	return bearer, nil
}
