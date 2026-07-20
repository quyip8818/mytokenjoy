package apply

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func insertSeedOrgNodes(ctx context.Context, exec TableWriter, tid uuid.UUID, nodes []types.OrgNode) error {
	paths := store.ComputeOrgNodePaths(nodes)
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	for i, node := range flat {
		path, ok := paths[node.ID]
		if !ok {
			path = store.OrgNodePathLabel(node.ID)
		}
		if err := insertSeedOrgNodeWithBudget(ctx, exec, tid, node, path, i); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedOrgNodeWithBudget(ctx context.Context, exec TableWriter, tid uuid.UUID, node types.OrgNode, path string, sortOrder int) error {
	row := pkgbudget.OrgNodeBudgetRowFromNode(node)
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_nodes (
			id, company_id, name, parent_id, path, external_id, source, manager_id, sort_order,
			default_model_id, fallback_model_id, routing_inherited,
			budget, reserved_pool, period, member_avg_budget
		) VALUES ($1, $2, $3, $4, $5::ltree, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (company_id, id) DO UPDATE SET
			budget = EXCLUDED.budget,
			reserved_pool = EXCLUDED.reserved_pool,
			period = EXCLUDED.period,
			member_avg_budget = EXCLUDED.member_avg_budget,
			updated_at = NOW()
	`, node.ID, tid, node.Name, node.ParentID, path,
		node.ExternalID, node.Source, node.ManagerID, sortOrder,
		node.DefaultModelID, node.FallbackModelID, node.RoutingInherited,
		row.Budget, row.ReservedPool, row.Period, row.MemberAvgBudget); err != nil {
		return fmt.Errorf("seed org node %s: %w", node.ID, err)
	}
	return nil
}
