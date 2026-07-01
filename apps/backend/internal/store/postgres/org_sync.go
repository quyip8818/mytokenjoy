package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) SyncConfig(ctx context.Context) (types.SyncConfig, error) {
	companyID := store.CompanyID(ctx)
	var cfg types.SyncConfig
	err := r.db.QueryRow(ctx, `
		SELECT enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im
		FROM org_sync_config WHERE company_id = $1
	`, companyID).Scan(
		&cfg.Enabled, &cfg.StartTime, &cfg.FrequencyHours,
		&cfg.DeleteMemberThreshold, &cfg.DeleteDepartmentThreshold,
		&cfg.NotifyPhone, &cfg.NotifyEmail, &cfg.NotifyIm,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.SyncConfig{}, nil
		}
		return types.SyncConfig{}, err
	}
	return cfg, nil
}

func (r *pgOrgRepo) SetSyncConfig(ctx context.Context, cfg types.SyncConfig) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO org_sync_config (
			company_id, enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			start_time = EXCLUDED.start_time,
			frequency_hours = EXCLUDED.frequency_hours,
			delete_member_threshold = EXCLUDED.delete_member_threshold,
			delete_department_threshold = EXCLUDED.delete_department_threshold,
			notify_phone = EXCLUDED.notify_phone,
			notify_email = EXCLUDED.notify_email,
			notify_im = EXCLUDED.notify_im,
			updated_at = NOW()
	`, companyID, cfg.Enabled, cfg.StartTime, cfg.FrequencyHours,
		cfg.DeleteMemberThreshold, cfg.DeleteDepartmentThreshold,
		cfg.NotifyPhone, cfg.NotifyEmail, cfg.NotifyIm)
	if err != nil {
		return fmt.Errorf("upsert sync config: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) SyncLogs(ctx context.Context) ([]types.SyncLog, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, time, type, result, detail
		FROM org_sync_logs
		WHERE company_id = $1
		ORDER BY time DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.SyncLog, 0)
	for rows.Next() {
		var item types.SyncLog
		var t time.Time
		if err := rows.Scan(&item.ID, &t, &item.Type, &item.Result, &item.Detail); err != nil {
			return nil, err
		}
		item.Time = formatSyncLogTime(t)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneSyncLogs(items), nil
}

func (r *pgOrgRepo) AppendSyncLog(ctx context.Context, log types.SyncLog) error {
	companyID := store.CompanyID(ctx)
	t, err := parseAPITime(log.Time)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO org_sync_logs (id, company_id, time, type, result, detail)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_id, id) DO UPDATE SET
			time = EXCLUDED.time,
			type = EXCLUDED.type,
			result = EXCLUDED.result,
			detail = EXCLUDED.detail
	`, log.ID, companyID, t, log.Type, log.Result, log.Detail)
	if err != nil {
		return fmt.Errorf("append sync log: %w", err)
	}
	return nil
}
