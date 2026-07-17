package apply

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func insertSeedTenantBackgroundState(ctx context.Context, exec TableWriter, companyIDs ...uuid.UUID) error {
	for _, id := range companyIDs {
		if _, err := exec.Exec(ctx, `
			INSERT INTO tenant_background_state (company_id)
			VALUES ($1)
			ON CONFLICT (company_id) DO NOTHING
		`, id); err != nil {
			return fmt.Errorf("seed tenant_background_state %s: %w", id, err)
		}
	}
	return nil
}
