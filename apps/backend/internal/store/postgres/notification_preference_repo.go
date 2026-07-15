package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type notificationPreferenceRepo struct {
	db dbQuerier
}

func (r *notificationPreferenceRepo) Get(ctx context.Context, userID string) ([]types.NotificationPreferenceEntry, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT category, channel, enabled
		FROM notification_preferences
		WHERE company_id = $1 AND user_id = $2
		ORDER BY category, channel
	`, companyID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NotificationPreferenceEntry
	for rows.Next() {
		var e types.NotificationPreferenceEntry
		if err := rows.Scan(&e.Category, &e.Channel, &e.Enabled); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *notificationPreferenceRepo) Upsert(ctx context.Context, userID string, entries []types.NotificationPreferenceEntry) error {
	companyID := store.CompanyID(ctx)
	for _, e := range entries {
		_, err := r.db.Exec(ctx, `
			INSERT INTO notification_preferences (company_id, user_id, category, channel, enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, user_id, category, channel)
			DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = NOW()
		`, companyID, userID, e.Category, e.Channel, e.Enabled)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *notificationPreferenceRepo) Delete(ctx context.Context, userID string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		DELETE FROM notification_preferences
		WHERE company_id = $1 AND user_id = $2
	`, companyID, userID)
	return err
}

var _ store.NotificationPreferenceRepository = (*notificationPreferenceRepo)(nil)
