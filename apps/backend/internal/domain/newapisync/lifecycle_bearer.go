package newapisync

import (
	"context"
	"strings"

	"github.com/tokenjoy/backend/internal/domain"
)

func (l *NewAPISync) ResolvePlatformKeyBearer(ctx context.Context, platformKeyID string) (string, error) {
	if !l.Enabled() {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	_, mapping, err := l.requireSyncedMapping(ctx, platformKeyID)
	if err != nil {
		return "", err
	}
	bearer, err := l.client.GetTokenKey(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return "", err
	}
	if bearer == "" || strings.Contains(bearer, "*") {
		return "", domain.Conflict("platform key secret unavailable; rotate once via /keys/platform")
	}
	return bearer, nil
}
