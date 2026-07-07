package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgNodeRepo struct {
	db dbQuerier
}

func (r *pgOrgNodeRepo) Tree(ctx context.Context) ([]types.OrgNode, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, parent_id, member_count, external_id, source, manager_id, sort_order,
			budget, consumed, reserved_pool, period, default_model, fallback_model, routing_inherited
		FROM org_nodes
		WHERE company_id = $1
		ORDER BY sort_order
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flat := make([]flatOrgNode, 0)
	for rows.Next() {
		var row flatOrgNode
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.MemberCount,
			&row.ExternalID, &row.Source, &row.ManagerID, &row.sortOrder,
			&row.Budget, &row.Consumed, &row.ReservedPool, &row.Period,
			&row.DefaultModel, &row.FallbackModel, &row.RoutingInherited,
		); err != nil {
			return nil, err
		}
		flat = append(flat, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneOrgNodes(buildOrgNodeTree(flat)), nil
}

func (r *pgOrgNodeRepo) SetTree(ctx context.Context, tree []types.OrgNode) error {
	companyID := store.CompanyID(ctx)
	flat := flattenOrgNodesWithOrder(store.CloneOrgNodes(tree))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO org_nodes (
				id, company_id, name, parent_id, member_count, external_id, source, manager_id, sort_order,
				budget, consumed, reserved_pool, period, default_model, fallback_model, routing_inherited, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				member_count = EXCLUDED.member_count,
				external_id = EXCLUDED.external_id,
				source = EXCLUDED.source,
				manager_id = EXCLUDED.manager_id,
				sort_order = EXCLUDED.sort_order,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				reserved_pool = EXCLUDED.reserved_pool,
				period = EXCLUDED.period,
				default_model = EXCLUDED.default_model,
				fallback_model = EXCLUDED.fallback_model,
				routing_inherited = EXCLUDED.routing_inherited,
				updated_at = NOW()
		`, row.ID, companyID, row.Name, row.ParentID, row.MemberCount,
			row.ExternalID, row.Source, row.ManagerID, row.sortOrder,
			row.Budget, row.Consumed, row.ReservedPool, row.Period,
			row.DefaultModel, row.FallbackModel, row.RoutingInherited); err != nil {
			return fmt.Errorf("upsert org node %s: %w", row.ID, err)
		}
	}
	return pruneByIDForCompany(ctx, r.db, "org_nodes", companyID, ids)
}

func (r *pgOrgNodeRepo) RollupConsumed(ctx context.Context, nodeID string, amountCNY float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id FROM org_nodes
			WHERE company_id = $1 AND id = $2
			UNION ALL
			SELECT n.id, n.parent_id FROM org_nodes n
			INNER JOIN ancestors a ON n.id = a.parent_id
			WHERE n.company_id = $1 AND a.parent_id IS NOT NULL
		)
		UPDATE org_nodes SET consumed = consumed + $3, updated_at = NOW()
		WHERE company_id = $1 AND id IN (SELECT id FROM ancestors)
	`, companyID, nodeID, amountCNY)
	return err
}

func (r *pgOrgNodeRepo) GetNodeBudget(ctx context.Context, nodeID string) (float64, float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var budget, consumed float64
	err := r.db.QueryRow(ctx, `
		SELECT budget, consumed FROM org_nodes WHERE company_id = $1 AND id = $2
	`, companyID, nodeID).Scan(&budget, &consumed)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return budget, consumed, true, nil
}

type flatOrgNode struct {
	types.OrgNode
	sortOrder int
}

func flattenOrgNodesWithOrder(nodes []types.OrgNode) []flatOrgNode {
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	result := make([]flatOrgNode, len(flat))
	for i, node := range flat {
		result[i] = flatOrgNode{OrgNode: node, sortOrder: i}
	}
	return result
}

func buildOrgNodeTree(rows []flatOrgNode) []types.OrgNode {
	if len(rows) == 0 {
		return nil
	}
	byID := make(map[string]flatOrgNode, len(rows))
	orderByID := make(map[string]int, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
		orderByID[row.ID] = row.sortOrder
	}
	var build func(id string) types.OrgNode
	build = func(id string) types.OrgNode {
		row := byID[id]
		node := row.OrgNode
		node.Children = nil
		children := make([]types.OrgNode, 0)
		for _, child := range rows {
			if child.ParentID == nil || *child.ParentID != id {
				continue
			}
			children = append(children, build(child.ID))
		}
		sortOrgNodeSiblings(children, orderByID)
		node.Children = children
		return node
	}
	roots := make([]types.OrgNode, 0)
	for _, row := range rows {
		if row.ParentID != nil && *row.ParentID != "" {
			if _, ok := byID[*row.ParentID]; ok {
				continue
			}
		}
		roots = append(roots, build(row.ID))
	}
	sortOrgNodeSiblings(roots, orderByID)
	return roots
}

func sortOrgNodeSiblings(nodes []types.OrgNode, orderByID map[string]int) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if orderByID[nodes[i].ID] > orderByID[nodes[j].ID] {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
