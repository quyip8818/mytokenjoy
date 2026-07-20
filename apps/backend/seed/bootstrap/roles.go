package bootstrap

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/grants"
)

// seedGlobalPresetRoles 写入全局预设角色行。幂等。
func seedGlobalPresetRoles(ctx context.Context, exec TableWriter) error {
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
