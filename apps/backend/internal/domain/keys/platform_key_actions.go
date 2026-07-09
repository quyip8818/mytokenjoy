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
	if err := s.requireRelay(); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			targetActive := enabled
			if err := s.relaySync.SyncUpdatePlatformKey(ctx, id, &targetActive); err != nil {
				return types.PlatformKey{}, err
			}
			if enabled {
				platformKeys[i].Status = "active"
			} else {
				platformKeys[i].Status = "disabled"
			}
			if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return types.PlatformKey{}, err
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
	if err := s.requireRelay(); err != nil {
		return types.PlatformKey{}, err
	}
	fullKey, err := s.relaySync.SyncRotatePlatformKey(ctx, id)
	if err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			key := platformKeys[i]
			key.FullKey = &fullKey
			return s.enrichPlatformKeyResponse(ctx, key)
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RevokePlatformKey(ctx context.Context, id string) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	if err := s.requireRelay(); err != nil {
		return err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			if err := s.relaySync.SyncRevokePlatformKey(ctx, id); err != nil {
				return err
			}
			platformKeys[i].Status = "revoked"
			return s.store.Keys().SetPlatformKeys(ctx, platformKeys)
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
