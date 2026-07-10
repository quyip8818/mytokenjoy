package newapisync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) EnqueueUpsertProviderKey(ctx context.Context, providerKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpsertChannelOutboxPayload{
		CompanyID:     company.CompanyID(ctx),
		ProviderKeyID: providerKeyID,
	})
	return l.outbox.EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID:      fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:    store.OutboxKindUpsertChannel,
		Payload: payload,
		Status:  store.JobStatusPending,
	})
}

func (l *NewAPISync) SyncUpsertProviderKey(ctx context.Context, providerKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	keys, err := l.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return err
	}
	idx := -1
	for i := range keys {
		if keys[i].ID == providerKeyID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("provider key not found: %s", providerKeyID)
	}
	pk := keys[idx]
	if pk.SecretKey == "" {
		return fmt.Errorf("provider key secret missing: %s", providerKeyID)
	}
	status := newapi.ChannelStatusEnabled
	if pk.Status != "active" {
		status = newapi.ChannelStatusDisabled
	}
	req := newapi.UpsertChannelRequest{
		Type:   newapi.ProviderChannelType(pk.Provider),
		Name:   pk.Name,
		Key:    pk.SecretKey,
		Status: status,
	}
	if pk.NewAPIChannelID > 0 {
		req.ID = pk.NewAPIChannelID
	}
	channel, err := l.client.UpsertChannel(ctx, req)
	if err != nil {
		return err
	}
	keys[idx].NewAPIChannelID = channel.ID
	return l.store.Keys().SetProviderKeys(ctx, keys)
}
