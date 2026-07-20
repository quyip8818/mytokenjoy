package platformkey

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
)

func newAPIPlatformKeyPrefix(fullKey string) string {
	prefix := fullKey
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	return prefix
}

func persistPlatformKeySecret(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID, fullKey string) error {
	keys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == platformKeyID {
			keys[i].FullKey = &fullKey
			keys[i].KeyPrefix = newAPIPlatformKeyPrefix(fullKey)
			return d.Store.Keys().SetPlatformKeys(ctx, keys)
		}
	}
	return fmt.Errorf("platform key not found: %s", platformKeyID)
}
