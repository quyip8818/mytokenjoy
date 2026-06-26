package queryutil

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
)

func CollectDescendantDeptIDs(departments []types.Department, rootID string) []string {
	flat := orgutil.FlattenDepartmentTree(departments)
	childrenByParent := make(map[string][]string)
	for _, dept := range flat {
		if dept.ParentID != nil {
			parentID := *dept.ParentID
			childrenByParent[parentID] = append(childrenByParent[parentID], dept.ID)
		}
	}

	result := make([]string, 0)
	queue := []string{rootID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, childrenByParent[current]...)
	}
	return result
}

func FilterMembersByDepartment(
	members []types.Member,
	departments []types.Department,
	departmentID string,
	directOnly bool,
) []types.Member {
	if directOnly {
		filtered := make([]types.Member, 0)
		for _, member := range members {
			if member.DepartmentID == departmentID {
				filtered = append(filtered, member)
			}
		}
		return filtered
	}

	allowed := make(map[string]struct{})
	for _, id := range CollectDescendantDeptIDs(departments, departmentID) {
		allowed[id] = struct{}{}
	}

	filtered := make([]types.Member, 0)
	for _, member := range members {
		if _, ok := allowed[member.DepartmentID]; ok {
			filtered = append(filtered, member)
		}
	}
	return filtered
}

func FindMemberByID(members []types.Member, memberID string) (*types.Member, bool) {
	for i := range members {
		if members[i].ID == memberID {
			return &members[i], true
		}
	}
	return nil, false
}
