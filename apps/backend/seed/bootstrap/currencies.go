package bootstrap

import (
	"context"
	"fmt"
)

func insertCurrencies(ctx context.Context, exec TableWriter, cfg Config) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (currency) DO NOTHING
	`, cfg.Billing.Currency, cfg.Billing.QuotaPerUnit); err != nil {
		return fmt.Errorf("insert currency %s: %w", cfg.Billing.Currency, err)
	}
	return nil
}
