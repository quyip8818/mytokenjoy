package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type notificationRepo struct {
	db dbQuerier
}

func (r *notificationRepo) Append(ctx context.Context, entry types.NotificationLogEntry) error {
	companyID := store.CompanyID(ctx)
	var userID *uuid.UUID
	if entry.UserID != uuid.Nil {
		userID = &entry.UserID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_log (id, company_id, channel, event_type, user_id, title, body, payload, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), NOW())
	`, entry.ID, companyID, entry.Channel, entry.EventType, userID, entry.Title, entry.Body, entry.Payload, entry.Status, entry.Error)
	return err
}

func (r *notificationRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]types.NotificationLogEntry, error) {
	companyID := store.CompanyID(ctx)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, channel, event_type, user_id, title, body, payload, status, COALESCE(error,''), created_at, read_at
		FROM notification_log
		WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, companyID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NotificationLogEntry
	for rows.Next() {
		var e types.NotificationLogEntry
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &e.UserID, &e.Title, &e.Body, &e.Payload, &e.Status, &e.Error, &e.CreatedAt, &e.ReadAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	companyID := store.CompanyID(ctx)
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notification_log
		WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app' AND read_at IS NULL
	`, companyID, userID).Scan(&count)
	return count, err
}

func (r *notificationRepo) MarkRead(ctx context.Context, id string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET read_at = $1, status = 'read'
		WHERE id = $2 AND company_id = $3 AND read_at IS NULL
	`, time.Now(), id, companyID)
	return err
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET read_at = $1, status = 'read'
		WHERE company_id = $2 AND user_id = $3 AND channel = 'in_app' AND read_at IS NULL
	`, time.Now(), companyID, userID)
	return err
}

var _ store.NotificationRepository = (*notificationRepo)(nil)

func (r *notificationRepo) ListLog(ctx context.Context, filter types.NotificationLogFilter) ([]types.NotificationLogEntry, error) {
	companyID := store.CompanyID(ctx)
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := `SELECT id, channel, event_type, COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid), title, body, payload, status, COALESCE(error,''), created_at, read_at
		FROM notification_log WHERE company_id = $1`
	args := []any{companyID}
	argIdx := 2

	if filter.Channel != "" {
		query += fmt.Sprintf(" AND channel = $%d", argIdx)
		args = append(args, filter.Channel)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.EventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, filter.EventType)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NotificationLogEntry
	for rows.Next() {
		var e types.NotificationLogEntry
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &e.UserID, &e.Title, &e.Body, &e.Payload, &e.Status, &e.Error, &e.CreatedAt, &e.ReadAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *notificationRepo) Stats(ctx context.Context) ([]types.NotificationStatRow, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT channel, status, COUNT(*) as cnt
		FROM notification_log
		WHERE company_id = $1
		GROUP BY channel, status
		ORDER BY channel, status
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NotificationStatRow
	for rows.Next() {
		var s types.NotificationStatRow
		if err := rows.Scan(&s.Channel, &s.Status, &s.Count); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
