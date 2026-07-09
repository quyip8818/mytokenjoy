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
	idx, ok := platformKeyIndex(platformKeys, id)
	if !ok {
		return types.PlatformKey{}, domain.NotFound("Not found")
	}
	targetActive := enabled
	if err := s.relaySync.SyncUpdatePlatformKey(ctx, id, &targetActive); err != nil {
		return types.PlatformKey{}, err
	}
	if enabled {
		platformKeys[idx].Status = "active"
	} else {
		platformKeys[idx].Status = "disabled"
	}
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}
	return s.enrichPlatformKeyResponse(ctx, platformKeys[idx])
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
	platformKeys, idx, err := s.relayRevokeKey(ctx, id)
	if err != nil {
		return err
	}
	platformKeys[idx].Status = "revoked"
	return s.store.Keys().SetPlatformKeys(ctx, platformKeys)
}

func (s *service) DeletePlatformKey(ctx context.Context, id string) error {
	platformKeys, idx, err := s.relayRevokeKey(ctx, id)
	if err != nil {
		return err
	}
	platformKeys = append(platformKeys[:idx], platformKeys[idx+1:]...)
	return s.store.Keys().SetPlatformKeys(ctx, platformKeys)
}
