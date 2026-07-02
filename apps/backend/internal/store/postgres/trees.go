package postgres

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

type flatDepartment struct {
	types.Department
	sortOrder int
}

func flattenDepartmentsWithOrder(departments []types.Department) []flatDepartment {
	flat := pkgorg.FlattenDepartmentTree(departments)
	result := make([]flatDepartment, len(flat))
	for i, dept := range flat {
		cloned := dept
		cloned.Children = nil
		result[i] = flatDepartment{Department: cloned, sortOrder: i}
	}
	return result
}

func buildDepartmentTree(rows []flatDepartment) []types.Department {
	if len(rows) == 0 {
		return nil
	}
	byID := make(map[string]*types.Department, len(rows))
	order := make([]flatDepartment, len(rows))
	copy(order, rows)
	for i := range order {
		dept := order[i].Department
		dept.Children = nil
		byID[dept.ID] = &order[i].Department
	}
	roots := make([]types.Department, 0)
	for _, row := range order {
		dept := byID[row.ID]
		if row.ParentID == nil || *row.ParentID == "" {
			roots = append(roots, *dept)
			continue
		}
		parent, ok := byID[*row.ParentID]
		if !ok {
			roots = append(roots, *dept)
			continue
		}
		parent.Children = append(parent.Children, *dept)
	}
	sortTreeChildren(roots, rows)
	return roots
}

func sortTreeChildren(nodes []types.Department, rows []flatDepartment) {
	orderByID := make(map[string]int, len(rows))
	for _, row := range rows {
		orderByID[row.ID] = row.sortOrder
	}
	sortDeptSiblings(nodes, orderByID)
	for i := range nodes {
		if len(nodes[i].Children) > 0 {
			sortTreeChildren(nodes[i].Children, rows)
		}
	}
}

func sortDeptSiblings(nodes []types.Department, orderByID map[string]int) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if orderByID[nodes[i].ID] > orderByID[nodes[j].ID] {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

type flatBudgetNode struct {
	types.BudgetNode
	sortOrder int
}

func flattenBudgetNodesWithOrder(nodes []types.BudgetNode) []flatBudgetNode {
	flat := pkgbudget.FlattenBudgetTree(nodes)
	result := make([]flatBudgetNode, len(flat))
	for i, node := range flat {
		cloned := node
		cloned.Children = nil
		result[i] = flatBudgetNode{BudgetNode: cloned, sortOrder: i}
	}
	return result
}

func buildBudgetTree(rows []flatBudgetNode) []types.BudgetNode {
	if len(rows) == 0 {
		return nil
	}
	byID := make(map[string]*types.BudgetNode, len(rows))
	order := make([]flatBudgetNode, len(rows))
	copy(order, rows)
	for i := range order {
		node := order[i].BudgetNode
		node.Children = nil
		byID[node.ID] = &order[i].BudgetNode
	}
	roots := make([]types.BudgetNode, 0)
	for _, row := range order {
		node := byID[row.ID]
		if row.ParentID == nil || *row.ParentID == "" {
			roots = append(roots, *node)
			continue
		}
		parent, ok := byID[*row.ParentID]
		if !ok {
			roots = append(roots, *node)
			continue
		}
		parent.Children = append(parent.Children, *node)
	}
	sortBudgetTreeChildren(roots, rows)
	return roots
}

func sortBudgetTreeChildren(nodes []types.BudgetNode, rows []flatBudgetNode) {
	orderByID := make(map[string]int, len(rows))
	for _, row := range rows {
		orderByID[row.ID] = row.sortOrder
	}
	sortBudgetSiblings(nodes, orderByID)
	for i := range nodes {
		if len(nodes[i].Children) > 0 {
			sortBudgetTreeChildren(nodes[i].Children, rows)
		}
	}
}

func sortBudgetSiblings(nodes []types.BudgetNode, orderByID map[string]int) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if orderByID[nodes[i].ID] > orderByID[nodes[j].ID] {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
