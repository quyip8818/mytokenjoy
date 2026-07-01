package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			if enabled {
				platformKeys[i].Status = "active"
			} else {
				platformKeys[i].Status = "disabled"
			}
			if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return types.PlatformKey{}, err
			}
			if s.lifecycle != nil && s.lifecycle.Enabled() {
				if err := s.lifecycle.EnqueueUpdatePlatformKey(ctx, id); err != nil {
					return types.PlatformKey{}, err
				}
			}
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			fullKey := fmt.Sprintf("tj-rot-%d-demo-secret", time.Now().UnixMilli())
			platformKeys[i].FullKey = &fullKey
			prefix := fullKey
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			platformKeys[i].KeyPrefix = prefix
			if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return types.PlatformKey{}, err
			}
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RevokePlatformKey(ctx context.Context, id string) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			platformKeys[i].Status = "revoked"
			if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return err
			}
			if s.lifecycle != nil && s.lifecycle.Enabled() {
				return s.lifecycle.SyncRevokePlatformKey(ctx, id)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) DeletePlatformKey(ctx context.Context, id string) error {
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			platformKeys = append(platformKeys[:i], platformKeys[i+1:]...)
			if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return err
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}
