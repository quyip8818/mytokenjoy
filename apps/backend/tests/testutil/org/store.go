package orgfix

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func LoadBudgetTreeT(t *testing.T, ctx context.Context, st store.Store) []types.BudgetNode {
	t.Helper()
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func PersistBudgetTreeT(t *testing.T, ctx context.Context, st store.Store, tree []types.BudgetNode) {
	t.Helper()
	if err := PersistBudgetTree(ctx, st, tree); err != nil {
		t.Fatal(err)
	}
}

func PersistBudgetTree(ctx context.Context, st store.Store, tree []types.BudgetNode) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	types.ApplyBudgetTreeToOrgNodes(nodes, tree)
	return st.Budget().OrgNodeBudget().UpsertMany(ctx, pkgbudget.OrgNodeBudgetRowsFromNodes(nodes))
}

func LoadDepartmentsT(t *testing.T, ctx context.Context, st store.Store) []types.Department {
	t.Helper()
	depts, err := common.LoadDepartments(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	return depts
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

func flattenDepartments(departments []types.Department) map[string]types.Department {
	result := make(map[string]types.Department)
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

func applyDepartmentsToOrgNodes(nodes []types.OrgNode, byID map[string]types.Department) {
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
