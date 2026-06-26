package orgutil

import (
	"github.com/tokenjoy/backend/internal/domain/types"
)

func FlattenDepartmentTree(departments []types.Department) []types.Department {
	result := make([]types.Department, 0)
	for _, dept := range departments {
		result = append(result, dept)
		if len(dept.Children) > 0 {
			result = append(result, FlattenDepartmentTree(dept.Children)...)
		}
	}
	return result
}

func GetDeptPath(departments []types.Department, targetID string) *string {
	var walk func(nodes []types.Department, path []string) *string
	walk = func(nodes []types.Department, path []string) *string {
		for _, node := range nodes {
			current := append(path, node.Name)
			if node.ID == targetID {
				joined := joinPath(current)
				return &joined
			}
			if len(node.Children) > 0 {
				if found := walk(node.Children, current); found != nil {
					return found
				}
			}
		}
		return nil
	}
	return walk(departments, nil)
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " / " + parts[i]
	}
	return result
}
