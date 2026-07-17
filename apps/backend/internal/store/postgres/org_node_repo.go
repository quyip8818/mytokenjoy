package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgNodeRepo struct {
	db dbQuerier
}

func (r *pgOrgNodeRepo) Tree(ctx context.Context) ([]types.OrgNode, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT n.id, n.name, n.parent_id, n.external_id, n.source, n.manager_id, n.sort_order,
			COALESCE(b.budget, 0), b.reserved_pool,
			COALESCE(b.period, 'monthly'), n.default_model_id, n.fallback_model_id, n.routing_inherited,
			COALESCE(b.member_avg_budget, 0)
		FROM org_nodes n
		LEFT JOIN org_node_budget b ON b.company_id = n.company_id AND b.node_id = n.id
		WHERE n.company_id = $1
		ORDER BY n.sort_order
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flat := make([]flatOrgNode, 0)
	for rows.Next() {
		var row flatOrgNode
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID,
			&row.ExternalID, &row.Source, &row.ManagerID, &row.sortOrder,
			&row.Budget, &row.ReservedPool, &row.Period,
			&row.DefaultModelID, &row.FallbackModelID, &row.RoutingInherited,
			&row.MemberAvgBudget,
		); err != nil {
			return nil, err
		}
		row.MemberCount = 0
		row.Consumed = 0
		flat = append(flat, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tree := buildOrgNodeTree(flat)
	memberRows, err := r.db.Query(ctx, `
		SELECT department_id FROM members WHERE company_id = $1
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()
	members := make([]types.Member, 0)
	for memberRows.Next() {
		var member types.Member
		if err := memberRows.Scan(&member.DepartmentID); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := memberRows.Err(); err != nil {
		return nil, err
	}
	return pkgorg.RecalcOrgNodeMemberCounts(tree, members), nil
}

func (r *pgOrgNodeRepo) SetTree(ctx context.Context, tree []types.OrgNode) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneOrgNodes(tree)
	flat := flattenOrgNodesWithOrder(cloned)
	paths := store.ComputeOrgNodePaths(cloned)
	ids := make([]uuid.UUID, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		path, ok := paths[row.ID]
		if !ok {
			path = store.OrgNodePathLabel(row.ID)
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO org_nodes (
				id, company_id, name, parent_id, path, external_id, source, manager_id, sort_order,
				default_model_id, fallback_model_id, routing_inherited, updated_at
			) VALUES ($1, $2, $3, $4, $5::ltree, $6, $7, $8, $9, $10, $11, $12, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				path = EXCLUDED.path,
				external_id = EXCLUDED.external_id,
				source = EXCLUDED.source,
				manager_id = EXCLUDED.manager_id,
				sort_order = EXCLUDED.sort_order,
				default_model_id = EXCLUDED.default_model_id,
				fallback_model_id = EXCLUDED.fallback_model_id,
				routing_inherited = EXCLUDED.routing_inherited,
				updated_at = NOW()
		`, row.ID, companyID, row.Name, row.ParentID, path,
			row.ExternalID, row.Source, row.ManagerID, row.sortOrder,
			row.DefaultModelID, row.FallbackModelID, row.RoutingInherited); err != nil {
			return fmt.Errorf("upsert org node %s: %w", row.ID, err)
		}
	}
	if err := pruneByIDForCompanyUUID(ctx, r.db, "org_nodes", companyID, ids); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO org_node_budget (company_id, node_id, budget, period, member_avg_budget, updated_at)
		SELECT company_id, id, 0, 'monthly', 0, NOW()
		FROM org_nodes WHERE company_id = $1
		ON CONFLICT (company_id, node_id) DO NOTHING
	`, companyID)
	return err
}

func (r *pgOrgNodeRepo) GetNodeBudget(ctx context.Context, nodeID uuid.UUID) (float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var budget float64
	err := r.db.QueryRow(ctx, `
		SELECT budget FROM org_node_budget WHERE company_id = $1 AND node_id = $2
	`, companyID, nodeID).Scan(&budget)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return budget, true, nil
}

func (r *pgOrgNodeRepo) GetNodePeriod(ctx context.Context, nodeID uuid.UUID) (string, bool, error) {
	companyID := store.CompanyID(ctx)
	var period string
	err := r.db.QueryRow(ctx, `
		SELECT period FROM org_node_budget WHERE company_id = $1 AND node_id = $2
	`, companyID, nodeID).Scan(&period)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgbudget.PeriodMonthly, false, nil
		}
		return "", false, err
	}
	return period, true, nil
}

func (r *pgOrgNodeRepo) ListSelfAndAncestorIDs(ctx context.Context, leafNodeID uuid.UUID) ([]uuid.UUID, error) {
	if leafNodeID == uuid.Nil {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT ancestor.id
		FROM org_nodes leaf
		JOIN org_nodes ancestor
		  ON ancestor.company_id = leaf.company_id
		 AND ancestor.path @> leaf.path
		WHERE leaf.company_id = $1 AND leaf.id = $2
		ORDER BY nlevel(ancestor.path) ASC
	`, companyID, leafNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
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
	byID := make(map[uuid.UUID]flatOrgNode, len(rows))
	orderByID := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
		orderByID[row.ID] = row.sortOrder
	}
	var build func(id uuid.UUID) types.OrgNode
	build = func(id uuid.UUID) types.OrgNode {
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
		if row.ParentID != nil && *row.ParentID != uuid.Nil {
			if _, ok := byID[*row.ParentID]; ok {
				continue
			}
		}
		roots = append(roots, build(row.ID))
	}
	sortOrgNodeSiblings(roots, orderByID)
	return roots
}

func sortOrgNodeSiblings(nodes []types.OrgNode, orderByID map[uuid.UUID]int) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if orderByID[nodes[i].ID] > orderByID[nodes[j].ID] {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
