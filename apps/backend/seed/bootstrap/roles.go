package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/grants"
)

// seedGlobalPresetRoles 为给定公司写入预设角色行。幂等。
func seedGlobalPresetRoles(ctx context.Context, exec TableWriter, companyIDs ...uuid.UUID) error {
	for _, companyID := range companyIDs {
		for name, id := range grants.PresetRoles {
			if _, err := exec.Exec(ctx, `
				INSERT INTO roles (id, company_id, name, type, permissions)
				VALUES ($1, $2, $3, 'preset', '{}')
				ON CONFLICT (company_id, name) DO NOTHING
			`, id, companyID, name); err != nil {
				return fmt.Errorf("seed preset role %s for company %s: %w", name, companyID, err)
			}
		}
	}
	return nil
}
