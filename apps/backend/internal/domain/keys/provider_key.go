package keys

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error) {
	if s.cfg.SupportSaas {
		return types.ProviderKey{}, domain.Forbidden("provider keys are managed by platform in SaaS mode")
	}
	return s.createProviderKey(ctx, input)
}

func (s *service) CreateProviderKeyForPlatform(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error) {
	return s.createProviderKey(ctx, input)
}

func (s *service) createProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.ProviderKey{}, err
	}
	prefix := input.Key
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	created := types.ProviderKey{
		ID:            uuid.Must(uuid.NewV7()),
		Provider:      input.Provider,
		Name:          input.Name,
		KeyPrefix:     prefix,
		Status:        "active",
		CreatedAt:     s.cfg.SeedReferenceDate(),
		RotateEnabled: true,
		SecretKey:     input.Key,
	}
	keys, err := s.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return types.ProviderKey{}, err
	}
	keys = append(keys, created)
	if err := s.store.Keys().SetProviderKeys(ctx, keys); err != nil {
		return types.ProviderKey{}, err
	}
	if s.newAPISync != nil && s.newAPISync.Enabled() {
		if err := s.newAPISync.EnqueueUpsertProviderKey(ctx, created.ID); err != nil {
			return types.ProviderKey{}, err
		}
		if err := s.newAPISync.SyncUpsertProviderKey(ctx, created.ID); err != nil {
			return types.ProviderKey{}, domain.ServiceUnavailable("NewAPI Channel sync failed")
		}
	}
	return created, nil
}

func (s *service) ToggleProviderKey(ctx context.Context, id uuid.UUID, enabled bool) error {
	if s.cfg.SupportSaas {
		return domain.Forbidden("provider keys are managed by platform in SaaS mode")
	}
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	keys, err := s.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == id {
			if enabled {
				keys[i].Status = "active"
			} else {
				keys[i].Status = "disabled"
			}
			if err := s.store.Keys().SetProviderKeys(ctx, keys); err != nil {
				return err
			}
			if s.newAPISync != nil && s.newAPISync.Enabled() {
				if err := s.newAPISync.EnqueueUpsertProviderKey(ctx, id); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) RotateProviderKey(ctx context.Context, id uuid.UUID, newKey string) (types.ProviderKey, error) {
	if s.cfg.SupportSaas {
		return types.ProviderKey{}, domain.Forbidden("provider keys are managed by platform in SaaS mode")
	}
	if newKey == "" {
		return types.ProviderKey{}, domain.BadRequest("newKey is required")
	}
	if err := s.delayer.Wait(ctx, time.Second); err != nil {
		return types.ProviderKey{}, err
	}
	keys, err := s.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return types.ProviderKey{}, err
	}
	for i := range keys {
		if keys[i].ID == id {
			if !keys[i].RotateEnabled {
				return types.ProviderKey{}, domain.Forbidden("key rotation is disabled")
			}
			prefix := newKey
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			keys[i].SecretKey = newKey
			keys[i].KeyPrefix = prefix
			if err := s.store.Keys().SetProviderKeys(ctx, keys); err != nil {
				return types.ProviderKey{}, err
			}
			if s.newAPISync != nil && s.newAPISync.Enabled() {
				if err := s.newAPISync.EnqueueUpsertProviderKey(ctx, id); err != nil {
					return types.ProviderKey{}, err
				}
				if err := s.newAPISync.SyncUpsertProviderKey(ctx, id); err != nil {
					return types.ProviderKey{}, domain.ServiceUnavailable("NewAPI Channel sync failed")
				}
			}
			return keys[i], nil
		}
	}
	return types.ProviderKey{}, domain.NotFound("Not found")
}

func (s *service) DeleteProviderKey(ctx context.Context, id uuid.UUID) error {
	if s.cfg.SupportSaas {
		return domain.Forbidden("provider keys are managed by platform in SaaS mode")
	}
	keys, err := s.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == id {
			keys = append(keys[:i], keys[i+1:]...)
			if err := s.store.Keys().SetProviderKeys(ctx, keys); err != nil {
				return err
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}
