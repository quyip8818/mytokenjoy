package postgres

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) SyncConfig() types.SyncConfig {
	var cfg types.SyncConfig
	err := r.db.QueryRow(r.ctx, `
		SELECT enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im
		FROM org_sync_config WHERE id = 1
	`).Scan(
		&cfg.Enabled, &cfg.StartTime, &cfg.FrequencyHours,
		&cfg.DeleteMemberThreshold, &cfg.DeleteDepartmentThreshold,
		&cfg.NotifyPhone, &cfg.NotifyEmail, &cfg.NotifyIm,
	)
	if err != nil {
		return types.SyncConfig{}
	}
	return cfg
}

func (r *pgOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO org_sync_config (
			id, enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im, updated_at
		) VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			start_time = EXCLUDED.start_time,
			frequency_hours = EXCLUDED.frequency_hours,
			delete_member_threshold = EXCLUDED.delete_member_threshold,
			delete_department_threshold = EXCLUDED.delete_department_threshold,
			notify_phone = EXCLUDED.notify_phone,
			notify_email = EXCLUDED.notify_email,
			notify_im = EXCLUDED.notify_im,
			updated_at = NOW()
	`, cfg.Enabled, cfg.StartTime, cfg.FrequencyHours,
		cfg.DeleteMemberThreshold, cfg.DeleteDepartmentThreshold,
		cfg.NotifyPhone, cfg.NotifyEmail, cfg.NotifyIm)
	if err != nil {
		return fmt.Errorf("upsert sync config: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) SyncLogs() []types.SyncLog {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, time, type, result, detail
		FROM org_sync_logs
		ORDER BY time DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.SyncLog, 0)
	for rows.Next() {
		var item types.SyncLog
		var t time.Time
		if err := rows.Scan(&item.ID, &t, &item.Type, &item.Result, &item.Detail); err != nil {
			return nil
		}
		item.Time = formatSyncLogTime(t)
		items = append(items, item)
	}
	return store.CloneSyncLogs(items)
}

func (r *pgOrgRepo) AppendSyncLog(log types.SyncLog) error {
	t, err := parseAPITime(log.Time)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(r.ctx, `
		INSERT INTO org_sync_logs (id, time, type, result, detail)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			time = EXCLUDED.time,
			type = EXCLUDED.type,
			result = EXCLUDED.result,
			detail = EXCLUDED.detail
	`, log.ID, t, log.Type, log.Result, log.Detail)
	if err != nil {
		return fmt.Errorf("append sync log: %w", err)
	}
	return nil
}
