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
	createdAt := time.Now()
	if !entry.CreatedAt.IsZero() {
		createdAt = entry.CreatedAt
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_log (id, company_id, channel, event_type, user_id, title, body, payload, send_ok, error, category, group_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12, 'active', $13, $13)
	`, entry.ID, companyID, entry.Channel, entry.EventType, userID, entry.Title, entry.Body, entry.Payload, entry.SendOK, entry.Error, entry.Category, entry.GroupKey, createdAt)
	return err
}

func (r *notificationRepo) List(ctx context.Context, filter types.NotificationListFilter) (types.NotificationListResult, error) {
	companyID := store.CompanyID(ctx)
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if filter.Grouped && filter.GroupKey == "" {
		return r.listGrouped(ctx, companyID, filter, limit)
	}
	return r.listFlat(ctx, companyID, filter, limit)
}

func (r *notificationRepo) listFlat(ctx context.Context, companyID uuid.UUID, filter types.NotificationListFilter, limit int) (types.NotificationListResult, error) {
	query := `SELECT id, channel, event_type, user_id, title, body, payload, send_ok, COALESCE(error,''), category, group_key, status, created_at, read_at, updated_at
		FROM notification_log WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app' AND status != 'deleted'`
	args := []any{companyID, filter.UserID}
	argIdx := 3

	if filter.Archived {
		query += " AND status = 'archived'"
	} else {
		query += " AND status = 'active'"
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}

	if filter.Status == "unread" {
		query += " AND read_at IS NULL"
	} else if filter.Status == "read" {
		query += " AND read_at IS NOT NULL"
	}

	if filter.GroupKey != "" {
		query += fmt.Sprintf(" AND group_key = $%d", argIdx)
		args = append(args, filter.GroupKey)
		argIdx++
	}

	if filter.Cursor != nil {
		query += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *filter.Cursor)
		argIdx++
	}

	// Fetch limit+1 to know if there's more
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return types.NotificationListResult{}, err
	}
	defer rows.Close()

	var result []types.NotificationLogEntry
	for rows.Next() {
		var e types.NotificationLogEntry
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &e.UserID, &e.Title, &e.Body, &e.Payload, &e.SendOK, &e.Error, &e.Category, &e.GroupKey, &e.Status, &e.CreatedAt, &e.ReadAt, &e.UpdatedAt); err != nil {
			return types.NotificationListResult{}, err
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return types.NotificationListResult{}, err
	}

	hasMore := len(result) > limit
	if hasMore {
		result = result[:limit]
	}

	var nextCursor *time.Time
	if hasMore && len(result) > 0 {
		nextCursor = &result[len(result)-1].CreatedAt
	}

	return types.NotificationListResult{Items: result, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (r *notificationRepo) listGrouped(ctx context.Context, companyID uuid.UUID, filter types.NotificationListFilter, limit int) (types.NotificationListResult, error) {
	query := `
WITH group_heads AS (
    SELECT DISTINCT ON (COALESCE(NULLIF(group_key, ''), id::text))
        id, group_key, created_at
    FROM notification_log
    WHERE company_id = $1 AND user_id = $2
      AND channel = 'in_app'
      AND status != 'deleted'`

	args := []any{companyID, filter.UserID}
	argIdx := 3

	if filter.Archived {
		query += " AND status = 'archived'"
	} else {
		query += " AND status = 'active'"
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Status == "unread" {
		query += " AND read_at IS NULL"
	} else if filter.Status == "read" {
		query += " AND read_at IS NOT NULL"
	}

	query += `
    ORDER BY COALESCE(NULLIF(group_key, ''), id::text), created_at DESC
),
paged AS (
    SELECT id, group_key, created_at
    FROM group_heads`

	if filter.Cursor != nil {
		query += fmt.Sprintf(" WHERE created_at < $%d", argIdx)
		args = append(args, *filter.Cursor)
		argIdx++
	}

	query += fmt.Sprintf(`
    ORDER BY created_at DESC
    LIMIT $%d
)
SELECT n.id, n.channel, n.event_type, n.user_id, n.title, n.body, n.payload, n.send_ok, COALESCE(n.error,''),
       n.category, n.group_key, n.status, n.created_at, n.read_at, n.updated_at,
       (SELECT COUNT(*) FROM notification_log sub
        WHERE sub.company_id = $1 AND sub.user_id = $2
          AND sub.channel = 'in_app'
          AND sub.status = 'active'
          AND COALESCE(NULLIF(sub.group_key, ''), sub.id::text) = COALESCE(NULLIF(n.group_key, ''), n.id::text)
       ) as group_count
FROM paged p
JOIN notification_log n ON n.id = p.id
ORDER BY p.created_at DESC`, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return types.NotificationListResult{}, err
	}
	defer rows.Close()

	type entryWithCount struct {
		types.NotificationLogEntry
		GroupCount int
	}

	var entries []entryWithCount
	for rows.Next() {
		var e entryWithCount
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &e.UserID, &e.Title, &e.Body, &e.Payload, &e.SendOK, &e.Error,
			&e.Category, &e.GroupKey, &e.Status, &e.CreatedAt, &e.ReadAt, &e.UpdatedAt, &e.GroupCount); err != nil {
			return types.NotificationListResult{}, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return types.NotificationListResult{}, err
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	result := make([]types.NotificationLogEntry, len(entries))
	for i, e := range entries {
		result[i] = e.NotificationLogEntry
		result[i].GroupCount = e.GroupCount
	}

	var nextCursor *time.Time
	if hasMore && len(result) > 0 {
		nextCursor = &result[len(result)-1].CreatedAt
	}

	return types.NotificationListResult{Items: result, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	companyID := store.CompanyID(ctx)
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notification_log
		WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app'
		  AND read_at IS NULL AND status = 'active'
	`, companyID, userID).Scan(&count)
	return count, err
}

func (r *notificationRepo) MarkRead(ctx context.Context, id uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET read_at = $1, updated_at = $2
		WHERE id = $3 AND company_id = $4 AND channel = 'in_app' AND read_at IS NULL
	`, now, now, id, companyID)
	return err
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET read_at = $1, updated_at = $2
		WHERE company_id = $3 AND user_id = $4 AND channel = 'in_app'
		  AND read_at IS NULL AND status = 'active'
	`, now, now, companyID, userID)
	return err
}

func (r *notificationRepo) Archive(ctx context.Context, id uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET status = 'archived', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND channel = 'in_app'
		  AND status = 'active'
	`, id, companyID)
	return err
}

func (r *notificationRepo) Unarchive(ctx context.Context, id uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET status = 'active', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND channel = 'in_app'
		  AND status = 'archived'
	`, id, companyID)
	return err
}

func (r *notificationRepo) ArchiveAll(ctx context.Context, userID uuid.UUID, category string) error {
	companyID := store.CompanyID(ctx)
	if category != "" {
		_, err := r.db.Exec(ctx, `
			UPDATE notification_log SET status = 'archived', updated_at = NOW()
			WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app'
			  AND category = $3 AND status = 'active'
		`, companyID, userID, category)
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET status = 'archived', updated_at = NOW()
		WHERE company_id = $1 AND user_id = $2 AND channel = 'in_app'
		  AND status = 'active'
	`, companyID, userID)
	return err
}

func (r *notificationRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND channel = 'in_app' AND status != 'deleted'
	`, id, companyID)
	return err
}

func (r *notificationRepo) Undelete(ctx context.Context, id uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	// ponytail: undelete restores to active; if it was archived before deletion we lose that info — acceptable tradeoff
	_, err := r.db.Exec(ctx, `
		UPDATE notification_log SET status = 'active', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND channel = 'in_app' AND status = 'deleted'
	`, id, companyID)
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

	query := `SELECT id, channel, event_type, user_id, title, body, payload, send_ok, COALESCE(error,''), category, group_key, status, created_at, read_at, updated_at
		FROM notification_log WHERE company_id = $1`
	args := []any{companyID}
	argIdx := 2

	if filter.Channel != "" {
		query += fmt.Sprintf(" AND channel = $%d", argIdx)
		args = append(args, filter.Channel)
		argIdx++
	}
	if filter.SendOK != nil {
		query += fmt.Sprintf(" AND send_ok = $%d", argIdx)
		args = append(args, *filter.SendOK)
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
		var userID *uuid.UUID
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &userID, &e.Title, &e.Body, &e.Payload, &e.SendOK, &e.Error, &e.Category, &e.GroupKey, &e.Status, &e.CreatedAt, &e.ReadAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if userID != nil {
			e.UserID = *userID
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *notificationRepo) Stats(ctx context.Context) ([]types.NotificationStatRow, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT channel, send_ok, COUNT(*) as cnt
		FROM notification_log
		WHERE company_id = $1
		GROUP BY channel, send_ok
		ORDER BY channel, send_ok
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NotificationStatRow
	for rows.Next() {
		var s types.NotificationStatRow
		if err := rows.Scan(&s.Channel, &s.SendOK, &s.Count); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
