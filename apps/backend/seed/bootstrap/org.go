package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/contract"
)

func insertRootOrg(ctx context.Context, exec TableWriter, companyID uuid.UUID, appCfg config.Config, cfg Config) error {
	// SaaS mode provides its own full org tree via snapshot — skip bootstrap root.
	if appCfg.SupportSaas {
		return nil
	}

	rootID := contract.IDDept1
	pathLabel := strings.ReplaceAll(rootID.String(), "-", "_")
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_nodes (id, company_id, name, parent_id, path, type, sort_order)
		VALUES ($1, $2, $3, NULL, $4::ltree, 'dept', 0)
		ON CONFLICT (company_id, id) DO NOTHING
	`, rootID, companyID, cfg.Company.Name, pathLabel); err != nil {
		return fmt.Errorf("insert root org node: %w", err)
	}

	// ponytail: admin user/member creation is handled by setup server (local) or snapshot (SaaS).
	// No demo admin seeded here.
	return nil
}
