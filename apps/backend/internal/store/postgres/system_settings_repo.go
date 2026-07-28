package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type pgSystemSettingsRepo struct {
	db dbQuerier
}

func (r *pgSystemSettingsRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRow(ctx, `SELECT value FROM system_settings WHERE key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil // missing key → empty string (treat as "0" by caller)
	}
	if err != nil {
		return "", fmt.Errorf("system_settings get %q: %w", key, err)
	}
	return value, nil
}

func (r *pgSystemSettingsRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO system_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("system_settings set %q: %w", key, err)
	}
	return nil
}

var _ store.SystemSettingsRepository = (*pgSystemSettingsRepo)(nil)
