package orgfix

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func PersistBudgetTree(ctx context.Context, st store.Store, tree []types.BudgetNode) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	types.ApplyBudgetTreeToOrgNodes(nodes, tree)
	return st.Budget().OrgNodeBudget().UpsertMany(ctx, pkgbudget.OrgNodeBudgetRowsFromNodes(nodes))
}

func PersistDepartmentsT(t *testing.T, ctx context.Context, st store.Store, departments []types.Department) {
	t.Helper()
	if err := PersistDepartments(ctx, st, departments); err != nil {
		t.Fatal(err)
	}
}

func PersistDepartments(ctx context.Context, st store.Store, departments []types.Department) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	applyDepartmentsToOrgNodes(nodes, flattenDepartments(departments))
	return st.Org().Nodes().SetTree(ctx, nodes)
}

func flattenDepartments(departments []types.Department) map[uuid.UUID]types.Department {
	result := make(map[uuid.UUID]types.Department)
	var walk func([]types.Department)
	walk = func(items []types.Department) {
		for _, dept := range items {
			flat := dept
			flat.Children = nil
			result[dept.ID] = flat
			if len(dept.Children) > 0 {
				walk(dept.Children)
			}
		}
	}
	walk(departments)
	return result
}

func applyDepartmentsToOrgNodes(nodes []types.OrgNode, byID map[uuid.UUID]types.Department) {
	for i := range nodes {
		if dept, ok := byID[nodes[i].ID]; ok {
			nodes[i].Name = dept.Name
			nodes[i].ParentID = dept.ParentID
			nodes[i].MemberCount = dept.MemberCount
			nodes[i].ExternalID = dept.ExternalID
			nodes[i].Source = dept.Source
			nodes[i].ManagerID = dept.ManagerID
		}
		if len(nodes[i].Children) > 0 {
			applyDepartmentsToOrgNodes(nodes[i].Children, byID)
		}
	}
}
