package apply

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func insertSeedCurrencies(ctx context.Context, exec TableWriter) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (currency) DO NOTHING
	`, common.DefaultBillingCurrency, common.DefaultQuotaPerUnit); err != nil {
		return fmt.Errorf("seed currencies: %w", err)
	}
	return nil
}

func insertSeedCompany(ctx context.Context, exec TableWriter, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, name, status) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, contract.TokenJoyCompanyID, "TokenJoy", store.CompanyStatusActive); err != nil {
		return fmt.Errorf("seed tokenjoy company: %w", err)
	}
	t := snap.Company
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, name, status) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, t.ID, t.Name, t.Status); err != nil {
		return fmt.Errorf("seed company: %w", err)
	}
	return insertSeedTenantBackgroundState(ctx, exec, contract.TokenJoyCompanyID, t.ID)
}

func insertSeedPermissions(ctx context.Context, exec TableWriter, permissions []types.Permission) error {
	for _, perm := range permissions {
		if _, err := exec.Exec(ctx, `
			INSERT INTO permissions (id, name, grp) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, perm.ID, perm.Name, perm.Group); err != nil {
			return fmt.Errorf("seed permission %s: %w", perm.ID, err)
		}
	}
	return nil
}

func insertSeedRoles(ctx context.Context, exec TableWriter, roles []types.Role) error {
	for _, role := range roles {
		companyID := nilUUID(role.CompanyID)
		perms := role.Permissions
		if perms == nil {
			perms = []string{}
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, permissions)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type, permissions = EXCLUDED.permissions
		`, role.ID, companyID, role.Name, role.Type, perms); err != nil {
			return fmt.Errorf("seed role %s: %w", role.ID, err)
		}
	}
	return nil
}

// nilUUID returns nil if id is zero-value, otherwise returns &id.
func nilUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func buildSeedRoleNameIndex(roles []types.Role) map[string]uuid.UUID {
	index := make(map[string]uuid.UUID, len(roles))
	for _, role := range roles {
		index[role.Name] = role.ID
	}
	return index
}
