package filler

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/points"
)

type DeptBudget struct {
	DepartmentID   string
	DepartmentName string
	Count          int
}

var leafDeptBudgets = []DeptBudget{
	{DepartmentID: contract.IDDept3, DepartmentName: "后端组", Count: 8},
	{DepartmentID: contract.IDDept4, DepartmentName: "前端组", Count: 7},
	{DepartmentID: "dept-5", DepartmentName: "测试组", Count: 6},
	{DepartmentID: "dept-6", DepartmentName: "产品部", Count: 6},
	{DepartmentID: "dept-7", DepartmentName: "市场部", Count: 6},
	{DepartmentID: "dept-8", DepartmentName: "行政部", Count: 7},
}

func anchorMembers() []types.Member {
	return []types.Member{
		{
			ID: contract.IDMemberAdmin, CompanyID: contract.DefaultCompanyID, Name: "管理员", Phone: "13800000001", Email: "admin@example.com",
			DepartmentID: "dept-1", DepartmentName: "总公司", Status: "active",
			Roles: []string{permission.RoleSuperAdmin}, Source: "manual",
		},
		{
			ID: contract.IDMember1, CompanyID: contract.DefaultCompanyID, Name: "张三", Phone: "13812341234", Email: "zhangsan@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "imported",
		},
		{
			ID: "m-2", CompanyID: contract.DefaultCompanyID, Name: "李四", Phone: "13912345678", Email: "lisi@example.com",
			DepartmentID: "dept-3", DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleOrgAdmin, permission.RoleBudgetApprover}, Source: "imported",
		},
		{
			ID: contract.IDMember3, CompanyID: contract.DefaultCompanyID, Name: "王五", Phone: "", Email: "wangwu@example.com",
			DepartmentID: "dept-3", DepartmentName: "后端组", Status: "pending",
			Roles: []string{permission.RoleMember}, Source: "invited",
		},
		{
			ID: "m-4", CompanyID: contract.DefaultCompanyID, Name: "赵六", Phone: "13712349876", Email: "zhaoliu@example.com",
			DepartmentID: "dept-4", DepartmentName: "前端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "manual",
		},
		{
			ID: "m-5", CompanyID: contract.DefaultCompanyID, Name: "钱七", Phone: "13612340000", Email: "qianqi@example.com",
			DepartmentID: "dept-4", DepartmentName: "前端组", Status: "inactive",
			Roles: []string{permission.RoleMember}, Source: "imported",
		},
		{
			ID: contract.IDMemberAuditor, CompanyID: contract.DefaultCompanyID, Name: "孙审计", Phone: "13512345678", Email: "sunaudit@example.com",
			DepartmentID: "dept-8", DepartmentName: "行政部", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAuditor}, Source: "manual",
		},
		{
			ID: contract.IDMemberPure, CompanyID: contract.DefaultCompanyID, Name: "周八", Phone: "13412345678", Email: "zhouba@example.com",
			DepartmentID: "dept-3", DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember}, Source: "manual",
		},
	}
}

func pickStatus(index int) string {
	mod := index % 100
	if mod < 7 {
		return "inactive"
	}
	if mod < 15 {
		return "pending"
	}
	return "active"
}

func pickSource(index int) string {
	mod := index % 10
	if mod == 0 {
		return "invited"
	}
	if mod <= 2 {
		return "manual"
	}
	return "imported"
}

func anchorsInDept(members []types.Member, deptID string) []types.Member {
	result := make([]types.Member, 0)
	for _, member := range members {
		if member.DepartmentID == deptID {
			result = append(result, member)
		}
	}
	return result
}

func buildGeneratedMember(id string, index int, departmentID, departmentName string) types.Member {
	name := buildChineseName(index)
	phone := ""
	if pickStatus(index) != "pending" {
		phone = buildPhone(index)
	}
	return types.Member{
		ID: id, CompanyID: contract.DefaultCompanyID, Name: name, Phone: phone, Email: buildEmail(index),
		DepartmentID: departmentID, DepartmentName: departmentName,
		Status: pickStatus(index), Roles: []string{permission.RoleMember}, Source: pickSource(index),
	}
}

func BuildAnchorMembers() []types.Member {
	members := append([]types.Member{}, anchorMembers()...)
	assignSpecialRoles(members)
	applyMemberPersonalBudgets(members)
	return members
}

func BuildMembers() []types.Member {
	members := BuildAnchorMembers()
	seq := 6

	for _, budget := range leafDeptBudgets {
		anchors := anchorsInDept(members, budget.DepartmentID)
		generatedCount := budget.Count - len(anchors)
		for i := 0; i < generatedCount; i++ {
			members = append(members, buildGeneratedMember(
				"m-"+itoa(seq), seq, budget.DepartmentID, budget.DepartmentName,
			))
			seq++
		}
	}

	assignSpecialRoles(members)
	applyMemberPersonalBudgets(members)
	return members
}

var anchorPersonalBudgets = map[string]float64{
	contract.IDMemberAdmin:   50000,
	contract.IDMember1:       10000,
	"m-2":                    15000,
	"m-4":                    12000,
	contract.IDMemberAuditor: 5000,
	contract.IDMemberPure:    3000,
}

func applyMemberPersonalBudgets(members []types.Member) {
	for i := range members {
		if amount, ok := anchorPersonalBudgets[members[i].ID]; ok {
			members[i].PersonalBudget = points.FromDisplay(amount)
			continue
		}
		members[i].PersonalBudget = points.FromDisplay(common.DefaultPersonalBudget)
	}
}

func assignSpecialRoles(members []types.Member) {
	for i := range members {
		if members[i].ID == "m-6" && members[i].DepartmentID == "dept-3" {
			if !org.ContainsRole(members[i].Roles, permission.RoleOrgAdmin) {
				members[i].Roles = append(members[i].Roles, permission.RoleOrgAdmin)
			}
		}
	}

	for i := range members {
		if members[i].DepartmentID == "dept-6" && members[i].ID != "m-2" {
			if !org.ContainsRole(members[i].Roles, permission.RoleBudgetApprover) {
				members[i].Roles = append(members[i].Roles, permission.RoleBudgetApprover)
				break
			}
		}
	}

	auditorCount := 0
	for i := range members {
		if members[i].DepartmentID == "dept-8" && members[i].Status == "active" && auditorCount < 3 {
			if !org.ContainsRole(members[i].Roles, permission.RoleAuditor) {
				members[i].Roles = append(members[i].Roles, permission.RoleAuditor)
			}
			auditorCount++
		}
	}

	apiCallerCount := 0
	for i := range members {
		if members[i].Status == "active" && !org.ContainsRole(members[i].Roles, permission.RoleSuperAdmin) && apiCallerCount < 50 {
			if !org.ContainsRole(members[i].Roles, permission.RoleAPICaller) {
				members[i].Roles = append(members[i].Roles, permission.RoleAPICaller)
			}
			apiCallerCount++
		}
	}
}
