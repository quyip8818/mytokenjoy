package keys

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	if err := s.requireNewAPI(); err != nil {
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
	if err := s.newAPISync.SyncUpdatePlatformKey(ctx, id, &targetActive); err != nil {
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
	s.cacheInvalidator.InvalidateByKeyID(id)
	if enabled {
		if err := domainbudget.RefreshPlatformKeyCombined(ctx, s.store, id, s.cfg.Clock(), nil); err != nil {
			return types.PlatformKey{}, err
		}
	}
	return s.enrichPlatformKeyResponse(ctx, platformKeys[idx])
}

func (s *service) RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error) {
	if err := s.requireNewAPI(); err != nil {
		return types.PlatformKey{}, err
	}
	fullKey, err := s.newAPISync.SyncRotatePlatformKey(ctx, id)
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
	platformKeys, idx, err := s.newAPIRevokeKey(ctx, id)
	if err != nil {
		return err
	}
	platformKeys[idx].Status = "revoked"
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return err
	}
	s.cacheInvalidator.InvalidateByKeyID(id)
	return nil
}

func (s *service) DeletePlatformKey(ctx context.Context, id string) error {
	platformKeys, idx, err := s.newAPIRevokeKey(ctx, id)
	if err != nil {
		return err
	}
	platformKeys = append(platformKeys[:idx], platformKeys[idx+1:]...)
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return err
	}
	s.cacheInvalidator.InvalidateByKeyID(id)
	return nil
}
