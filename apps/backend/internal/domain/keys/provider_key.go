package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.ProviderKey{}, err
	}
	prefix := input.Key
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	created := types.ProviderKey{
		ID:            fmt.Sprintf("pk-%d", time.Now().UnixMilli()),
		Provider:      input.Provider,
		Name:          input.Name,
		KeyPrefix:     prefix,
		Status:        "active",
		CreatedAt:     s.cfg.DemoToday,
		RotateEnabled: false,
		SecretKey:     input.Key,
	}
	keys := s.store.Keys().ProviderKeys()
	keys = append(keys, created)
	if err := s.store.Keys().SetProviderKeys(keys); err != nil {
		return types.ProviderKey{}, err
	}
	if s.lifecycle != nil && s.lifecycle.Enabled() {
		if err := s.lifecycle.EnqueueUpsertProviderKey(created.ID); err != nil {
			return types.ProviderKey{}, err
		}
		if err := s.lifecycle.SyncUpsertProviderKey(ctx, created.ID); err != nil {
			return types.ProviderKey{}, domain.NewDomainError(503, "Relay Channel 同步失败")
		}
	}
	return created, nil
}

func (s *service) ToggleProviderKey(ctx context.Context, id string) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	keys := s.store.Keys().ProviderKeys()
	for i := range keys {
		if keys[i].ID == id {
			if keys[i].Status == "active" {
				keys[i].Status = "disabled"
			} else {
				keys[i].Status = "active"
			}
			if err := s.store.Keys().SetProviderKeys(keys); err != nil {
				return err
			}
			if s.lifecycle != nil && s.lifecycle.Enabled() {
				_ = s.lifecycle.EnqueueUpsertProviderKey(id)
			}
			return nil
		}
	}
	return domain.NewDomainError(404, "Not found")
}

func (s *service) RotateProviderKey(ctx context.Context, id string) (types.ProviderKey, error) {
	if err := s.delayer.Wait(ctx, time.Second); err != nil {
		return types.ProviderKey{}, err
	}
	keys := s.store.Keys().ProviderKeys()
	for i := range keys {
		if keys[i].ID == id {
			keys[i].KeyPrefix = fmt.Sprintf("sk-rot-%x...", time.Now().UnixMilli())
			lastUsed := time.Now().Format("2006-01-02 15:04")
			keys[i].LastUsed = &lastUsed
			if err := s.store.Keys().SetProviderKeys(keys); err != nil {
				return types.ProviderKey{}, err
			}
			return keys[i], nil
		}
	}
	return types.ProviderKey{}, domain.NewDomainError(404, "Not found")
}

func (s *service) DeleteProviderKey(id string) error {
	_ = id
	return nil
}
