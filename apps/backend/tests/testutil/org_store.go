package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
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
	return st.Org().Nodes().SetTree(ctx, nodes)
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

func OrgNodesFromBudgetTree(tree []types.BudgetNode) []types.OrgNode {
	nodes := make([]types.OrgNode, len(tree))
	for i, node := range tree {
		nodes[i] = orgNodeFromBudgetNode(node)
	}
	return nodes
}

func orgNodeFromBudgetNode(node types.BudgetNode) types.OrgNode {
	out := types.OrgNode{
		ID:           node.ID,
		Name:         node.Name,
		ParentID:     node.ParentID,
		Budget:       node.Budget,
		Consumed:     node.Consumed,
		ReservedPool: node.ReservedPool,
		Period:       node.Period,
	}
	if len(node.Children) > 0 {
		out.Children = make([]types.OrgNode, len(node.Children))
		for i, child := range node.Children {
			out.Children[i] = orgNodeFromBudgetNode(child)
		}
	}
	return out
}
