package bootstrap

import (
	"context"
	"fmt"
)

func insertCurrencies(ctx context.Context, exec TableWriter, cfg Config) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled, updated_at)
		VALUES ($1, $2, TRUE, '2026-06-01T00:00:00Z')
	`, cfg.Billing.Currency, cfg.Billing.QuotaPerUnit); err != nil {
		return fmt.Errorf("insert currency %s: %w", cfg.Billing.Currency, err)
	}
	return nil
}
