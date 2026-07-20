package bootstrap

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/infra/permission"
)

func insertPermissions(ctx context.Context, exec TableWriter) error {
	manifest := permission.MustManifest()
	for id, meta := range manifest.PermissionNames {
		if _, err := exec.Exec(ctx, `
			INSERT INTO permissions (id, name, grp) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, id, meta.Name, meta.Group); err != nil {
			return fmt.Errorf("insert permission %s: %w", id, err)
		}
	}
	return nil
}
