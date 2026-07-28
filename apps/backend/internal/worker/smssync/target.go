package smssync

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/store"
)

// SyncTarget abstracts the write operations for SMS sync (NewAPI + local DB).
type SyncTarget interface {
	ReplaceChannels(ctx context.Context, channels []sms.CatalogChannel) error
	SyncModels(ctx context.Context, models []sms.CatalogModel) error
	ReplaceModelRatios(ctx context.Context, models []sms.CatalogModel) error
}

// Target implements SyncTarget using adminport.Port (channels/pricing) and store (models).
type Target struct {
	port            adminport.Port
	store           store.Store
	globalCompanyID uuid.UUID
}

func NewTarget(port adminport.Port, st store.Store, globalCompanyID uuid.UUID) *Target {
	return &Target{port: port, store: st, globalCompanyID: globalCompanyID}
}

// smsChannelPrefix is the naming convention for SMS-managed channels.
// Only channels with this prefix are candidates for diff-delete.
const smsChannelPrefix = "sms:"

// ReplaceChannels upserts all catalog channels, diff-deletes stale SMS-managed channels, rebuilds abilities.
func (t *Target) ReplaceChannels(ctx context.Context, channels []sms.CatalogChannel) error {
	if len(channels) == 0 {
		return nil // nothing to sync — skip to avoid accidental mass-delete
	}

	catalogNames := make(map[string]bool, len(channels))
	for _, ch := range channels {
		name := smsChannelPrefix + ch.Name
		_, err := t.port.UpsertChannel(ctx, adminport.UpsertChannelInput{
			Type:     ch.Type,
			Name:     name,
			Key:      ch.Key,
			Status:   1, // enabled
			Group:    ch.Group,
			BaseURL:  ch.BaseURL,
			Models:   strings.Join(ch.Models, ","),
			Priority: ch.Priority,
			Weight:   1,
			Settings: ch.Settings,
		})
		if err != nil {
			return fmt.Errorf("upsert channel %q: %w", name, err)
		}
		catalogNames[name] = true
	}

	// Diff delete: only remove SMS-prefixed channels not in the current catalog.
	existing, err := t.port.ListChannels(ctx)
	if err != nil {
		return fmt.Errorf("list channels for diff delete: %w", err)
	}
	for _, ex := range existing {
		if strings.HasPrefix(ex.Name, smsChannelPrefix) && !catalogNames[ex.Name] {
			_ = t.port.DeleteChannel(ctx, ex.ID) // best-effort
		}
	}

	return t.port.RebuildAbilities(ctx)
}

// SyncModels writes all SMS models to the global company (diff-disable + upsert).
func (t *Target) SyncModels(ctx context.Context, models []sms.CatalogModel) error {
	infos := make([]types.ModelInfo, 0, len(models))
	for _, m := range models {
		infos = append(infos, types.ModelInfo{
			CompanyID: t.globalCompanyID,
			Provider:  m.Provider,
			Type:      m.ModelID,
			Name:      m.DisplayName,
			Source:    "sms",
			Active:    true,
		})
	}
	return t.store.Models().SyncFromSMS(ctx, t.globalCompanyID, infos)
}

// ReplaceModelRatios batch-upserts model pricing into NewAPI (global, not per-company).
func (t *Target) ReplaceModelRatios(ctx context.Context, models []sms.CatalogModel) error {
	for _, m := range models {
		if err := t.port.UpsertModelRatio(ctx, m.ModelID, m.InputPrice, m.OutputPrice); err != nil {
			return fmt.Errorf("upsert model ratio %q: %w", m.ModelID, err)
		}
	}
	return nil
}
