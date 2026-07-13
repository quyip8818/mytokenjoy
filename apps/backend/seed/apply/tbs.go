package apply

import (
	"context"
	"fmt"
)

func insertSeedTenantBackgroundState(ctx context.Context, exec TableWriter, companyIDs ...int64) error {
	for _, id := range companyIDs {
		if _, err := exec.Exec(ctx, `
			INSERT INTO tenant_background_state (company_id)
			VALUES ($1)
			ON CONFLICT (company_id) DO NOTHING
		`, id); err != nil {
			return fmt.Errorf("seed tenant_background_state %d: %w", id, err)
		}
	}
	return nil
}
