package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type notificationRepo struct {
	db dbQuerier
}

func (r *notificationRepo) Append(ctx context.Context, entry types.NotificationLogEntry) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_log (id, company_id, channel, event_type, recipient, payload, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NOW())
	`, entry.ID, companyID, entry.Channel, entry.EventType, entry.Recipient, entry.Payload, entry.Status, entry.Error)
	return err
}

var _ store.NotificationRepository = (*notificationRepo)(nil)
