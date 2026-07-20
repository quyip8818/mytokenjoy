package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func insertPresetRoles(ctx context.Context, exec TableWriter, companyID uuid.UUID) error {
	manifest := permission.MustManifest()
	for roleName := range manifest.PresetRoles {
		roleID := grants.PresetRoleID(companyID, roleName)
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type) VALUES ($1, $2, $3, 'preset')
			ON CONFLICT (company_id, name) DO NOTHING
		`, roleID, companyID, roleName); err != nil {
			return fmt.Errorf("insert preset role %s: %w", roleName, err)
		}
	}
	return nil
}

// reconcileCompanyPresetRoles ensures a single company's preset role grants are up-to-date.
// Only adds grants declared in the manifest; never removes (respects admin customization).
func reconcileCompanyPresetRoles(ctx context.Context, exec TableWriter, companyID uuid.UUID) error {
	manifest := permission.MustManifest()
	allPermIDs := allPermissionIDs(manifest)

	for roleName, capabilities := range manifest.PresetRoles {
		roleID := grants.PresetRoleID(companyID, roleName)
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type) VALUES ($1, $2, $3, 'preset')
			ON CONFLICT (company_id, name) DO NOTHING
		`, roleID, companyID, roleName); err != nil {
			return fmt.Errorf("reconcile: ensure role %s: %w", roleName, err)
		}

		for _, permID := range resolveGrantIDs(capabilities, manifest, allPermIDs) {
			if _, err := exec.Exec(ctx, `
				INSERT INTO role_permission_grants (company_id, role_id, permission_id)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, companyID, roleID, permID); err != nil {
				return fmt.Errorf("reconcile: grant %s→%s: %w", roleName, permID, err)
			}
		}
	}
	return nil
}

// resolveGrantIDs expands capabilities (e.g. ["*"] or ["self:keys"]) to permission IDs.
func resolveGrantIDs(capabilities []string, manifest permission.Manifest, allPermIDs []string) []string {
	if len(capabilities) == 1 && capabilities[0] == "*" {
		return allPermIDs
	}
	capToID := make(map[string]string, len(manifest.PermissionIDMap))
	for id, cap := range manifest.PermissionIDMap {
		capToID[cap] = id
	}
	ids := make([]string, 0, len(capabilities))
	for _, cap := range capabilities {
		if id, ok := capToID[cap]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func allPermissionIDs(manifest permission.Manifest) []string {
	ids := make([]string, 0, len(manifest.PermissionIDMap))
	for id := range manifest.PermissionIDMap {
		ids = append(ids, id)
	}
	return ids
}
