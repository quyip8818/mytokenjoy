package smssync

import (
	"context"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/integration/sms"
)

// ModelStore abstracts the local model persistence for SMS-synced models.
type ModelStore interface {
	UpsertFromSMS(ctx context.Context, model sms.CatalogModel) error
}

// AdminPortTarget adapts adminport.Port + ModelStore into the SyncTarget interface.
type AdminPortTarget struct {
	port       adminport.Port
	modelStore ModelStore
}

func NewAdminPortTarget(port adminport.Port, modelStore ModelStore) *AdminPortTarget {
	return &AdminPortTarget{port: port, modelStore: modelStore}
}

func (t *AdminPortTarget) UpsertChannel(ctx context.Context, ch sms.CatalogChannel) error {
	_, err := t.port.UpsertChannel(ctx, adminport.UpsertChannelInput{
		Type:    ch.Type,
		Name:    ch.Name,
		Key:     ch.Key,
		Status:  1, // enabled
		Group:   ch.Group,
		BaseURL: ch.BaseURL,
		Models:  strings.Join(ch.Models, ","),
		Priority: ch.Priority,
		Weight:  1,
	})
	return err
}

func (t *AdminPortTarget) UpsertModelRatio(ctx context.Context, modelID string, inputPrice, outputPrice float64) error {
	return t.port.UpsertModelRatio(ctx, modelID, inputPrice, outputPrice)
}

func (t *AdminPortTarget) UpsertModel(ctx context.Context, model sms.CatalogModel) error {
	return t.modelStore.UpsertFromSMS(ctx, model)
}

func (t *AdminPortTarget) RebuildAbilities(ctx context.Context) error {
	return t.port.RebuildAbilities(ctx)
}
