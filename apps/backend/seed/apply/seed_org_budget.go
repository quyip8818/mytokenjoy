package apply

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func insertSeedOrgNodes(ctx context.Context, exec TableWriter, tid int64, nodes []types.OrgNode) error {
	paths := store.ComputeOrgNodePaths(nodes)
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	for i, node := range flat {
		path, ok := paths[node.ID]
		if !ok {
			path = store.OrgNodePathLabel(node.ID)
		}
		if err := insertSeedOrgNodeStructure(ctx, exec, tid, node, path, i); err != nil {
			return err
		}
		if err := insertSeedOrgNodeBudget(ctx, exec, tid, node); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedOrgNodeStructure(ctx context.Context, exec TableWriter, tid int64, node types.OrgNode, path string, sortOrder int) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_nodes (
			id, company_id, name, parent_id, path, external_id, source, manager_id, sort_order,
			default_model_id, fallback_model_id, routing_inherited
		) VALUES ($1, $2, $3, $4, $5::ltree, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (company_id, id) DO NOTHING
	`, node.ID, tid, node.Name, node.ParentID, path,
		node.ExternalID, node.Source, node.ManagerID, sortOrder,
		node.DefaultModelID, node.FallbackModelID,
		node.RoutingInherited); err != nil {
		return fmt.Errorf("seed org node %s: %w", node.ID, err)
	}
	return nil
}

func insertSeedOrgNodeBudget(ctx context.Context, exec TableWriter, tid int64, node types.OrgNode) error {
	row := pkgbudget.OrgNodeBudgetRowFromNode(node)
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_node_budget (
			company_id, node_id, budget, reserved_pool, period, member_avg_budget, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (company_id, node_id) DO UPDATE SET
			budget = EXCLUDED.budget,
			reserved_pool = EXCLUDED.reserved_pool,
			period = EXCLUDED.period,
			member_avg_budget = EXCLUDED.member_avg_budget,
			updated_at = NOW()
	`, tid, row.NodeID, row.Budget, row.ReservedPool, row.Period, row.MemberAvgBudget); err != nil {
		return fmt.Errorf("seed org node budget %s: %w", node.ID, err)
	}
	return nil
}
