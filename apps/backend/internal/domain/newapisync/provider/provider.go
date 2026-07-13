package provider

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
)

func EnqueueUpsertProviderKey(ctx context.Context, d syncdeps.Deps, providerKeyID string) error {
	if !syncdeps.Enabled(d) {
		return nil
	}
	return d.Enqueuer.InsertNewAPISync(ctx, ports.SyncJob{
		CompanyID:     company.CompanyID(ctx),
		SubKind:       outbox.KindUpsertChannel,
		ProviderKeyID: providerKeyID,
	})
}

func SyncUpsertProviderKey(ctx context.Context, d syncdeps.Deps, providerKeyID string) error {
	if !syncdeps.Enabled(d) {
		return nil
	}
	keys, err := d.Store.Keys().ProviderKeys(ctx)
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
	group := policy.ResolveProviderChannelGroup(d.Cfg)
	if err := d.Client.EnsureGroup(ctx, group, group); err != nil {
		return fmt.Errorf("ensure provider channel group %s: %w", group, err)
	}
	status := adminport.ChannelStatusEnabled
	if pk.Status != "active" {
		status = adminport.ChannelStatusDisabled
	}
	req := adminport.UpsertChannelInput{
		Type:   adminport.ProviderChannelType(pk.Provider),
		Name:   pk.Name,
		Key:    pk.SecretKey,
		Status: status,
		Group:  group,
	}
	if pk.NewAPIChannelID > 0 {
		req.ID = pk.NewAPIChannelID
	}
	channel, err := d.Client.UpsertChannel(ctx, req)
	if err != nil {
		return err
	}
	keys[idx].NewAPIChannelID = channel.ID
	return d.Store.Keys().SetProviderKeys(ctx, keys)
}
