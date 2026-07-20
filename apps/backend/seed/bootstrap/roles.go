package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/grants"
)

// seedGlobalPresetRoles inserts global preset roles (company_id = NULL). Idempotent.
func seedGlobalPresetRoles(ctx context.Context, exec TableWriter, _ ...uuid.UUID) error {
	for name, id := range grants.PresetRoles {
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, permissions)
			VALUES ($1, NULL, $2, 'preset', '{}')
			ON CONFLICT (id) DO NOTHING
		`, id, name); err != nil {
			return fmt.Errorf("seed preset role %s: %w", name, err)
		}
	}
	return nil
}
