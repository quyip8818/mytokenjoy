package keys

import (
	"context"
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
			if s.relaySync != nil && s.relaySync.Enabled() {
				if err := s.relaySync.EnqueueUpdatePlatformKey(ctx, id); err != nil {
					return types.PlatformKey{}, err
				}
			}
			enriched, err := s.enrichPlatformKeyResponse(ctx, platformKeys[i])
			if err != nil {
				return types.PlatformKey{}, err
			}
			return enriched, nil
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error) {
	_ = id
	return types.PlatformKey{}, domain.Unimplemented("Platform key rotation is not available")
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
			if s.relaySync != nil && s.relaySync.Enabled() {
				return s.relaySync.SyncRevokePlatformKey(ctx, id)
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
