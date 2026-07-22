package filler

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed/contract"
)

type DeptBudget struct {
	DepartmentID   uuid.UUID
	DepartmentName string
	Count          int
}

var leafDeptBudgets = []DeptBudget{
	{DepartmentID: contract.IDDept3, DepartmentName: "后端组", Count: 8},
	{DepartmentID: contract.IDDept4, DepartmentName: "前端组", Count: 7},
	{DepartmentID: contract.IDDept5, DepartmentName: "测试组", Count: 6},
	{DepartmentID: contract.IDDept6, DepartmentName: "产品部", Count: 6},
	{DepartmentID: contract.IDDept7, DepartmentName: "市场部", Count: 6},
	{DepartmentID: contract.IDDept8, DepartmentName: "行政部", Count: 7},
}

func anchorMembers() []types.Member {
	return []types.Member{
		{
			ID: contract.IDMemberAdmin, CompanyID: contract.DefaultCompanyID, Alias: "管理员", Phone: "13800000001", Email: "admin@example.com",
			DepartmentID: contract.IDDept1, DepartmentName: "总公司", Status: "active",
			Roles: []string{permission.RoleSuperAdmin}, Source: "manual",
		},
		{
			ID: contract.IDMember1, CompanyID: contract.DefaultCompanyID, Alias: "张三", Phone: "13812341234", Email: "zhangsan@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "imported",
		},
		{
			ID: contract.IDMember2, CompanyID: contract.DefaultCompanyID, Alias: "李四", Phone: "13912345678", Email: "lisi@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleOrgAdmin, permission.RoleBudgetApprover}, Source: "imported",
		},
		{
			ID: contract.IDMember3, CompanyID: contract.DefaultCompanyID, Alias: "王五", Phone: "", Email: "wangwu@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "pending",
			Roles: []string{permission.RoleMember}, Source: "invited",
		},
		{
			ID: contract.IDMember4, CompanyID: contract.DefaultCompanyID, Alias: "赵六", Phone: "13712349876", Email: "zhaoliu@example.com",
			DepartmentID: contract.IDDept4, DepartmentName: "前端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "manual",
		},
		{
			ID: contract.IDMember5, CompanyID: contract.DefaultCompanyID, Alias: "钱七", Phone: "13612340000", Email: "qianqi@example.com",
			DepartmentID: contract.IDDept4, DepartmentName: "前端组", Status: "inactive",
			Roles: []string{permission.RoleMember}, Source: "imported",
		},
		{
			ID: contract.IDMemberAuditor, CompanyID: contract.DefaultCompanyID, Alias: "孙审计", Phone: "13512345678", Email: "sunaudit@example.com",
			DepartmentID: contract.IDDept8, DepartmentName: "行政部", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAuditor}, Source: "manual",
		},
		{
			ID: contract.IDMemberPure, CompanyID: contract.DefaultCompanyID, Alias: "周八", Phone: "13412345678", Email: "zhouba@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember}, Source: "manual",
		},
		{
			ID: contract.IDMember6, CompanyID: contract.DefaultCompanyID, Alias: "吴九", Phone: "13312345678", Email: "wujiu@example.com",
			DepartmentID: contract.IDDept3, DepartmentName: "后端组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "imported",
		},
		{
			ID: contract.IDMember15, CompanyID: contract.DefaultCompanyID, Alias: "郑十五", Phone: "13212345615", Email: "zheng15@example.com",
			DepartmentID: contract.IDDept5, DepartmentName: "测试组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "imported",
		},
		{
			ID: contract.IDMember16, CompanyID: contract.DefaultCompanyID, Alias: "冯十六", Phone: "13212345616", Email: "feng16@example.com",
			DepartmentID: contract.IDDept5, DepartmentName: "测试组", Status: "active",
			Roles: []string{permission.RoleMember, permission.RoleAPICaller}, Source: "imported",
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

func anchorsInDept(members []types.Member, deptID uuid.UUID) []types.Member {
	result := make([]types.Member, 0)
	for _, member := range members {
		if member.DepartmentID == deptID {
			result = append(result, member)
		}
	}
	return result
}

// seedMemberID generates a stable UUID for generated members based on seq index.
func seedMemberID(seq int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-7000-8000-0000000e0%03x", seq))
}

func buildGeneratedMember(seq int, departmentID uuid.UUID, departmentName string) types.Member {
	name := buildChineseName(seq)
	phone := ""
	if pickStatus(seq) != "pending" {
		phone = buildPhone(seq)
	}
	return types.Member{
		ID: seedMemberID(seq), CompanyID: contract.DefaultCompanyID, Alias: name, Phone: phone, Email: buildEmail(seq),
		DepartmentID: departmentID, DepartmentName: departmentName,
		Status: pickStatus(seq), Roles: []string{permission.RoleMember}, Source: pickSource(seq),
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
			members = append(members, buildGeneratedMember(seq, budget.DepartmentID, budget.DepartmentName))
			seq++
		}
	}

	assignSpecialRoles(members)
	applyMemberPersonalBudgets(members)
	return members
}

var anchorPersonalBudgets = map[uuid.UUID]float64{
	contract.IDMemberAdmin:   50000,
	contract.IDMember1:       10000,
	contract.IDMember2:       15000,
	contract.IDMember4:       12000,
	contract.IDMemberAuditor: 5000,
	contract.IDMemberPure:    3000,
}

func applyMemberPersonalBudgets(members []types.Member) {
	for i := range members {
		if amount, ok := anchorPersonalBudgets[members[i].ID]; ok {
			members[i].PersonalBudget = common.QuotaFromAmount(amount, common.DefaultQuotaPerUnit)
			continue
		}
		members[i].PersonalBudget = common.DefaultPersonalBudget
	}
}

func assignSpecialRoles(members []types.Member) {
	for i := range members {
		if members[i].ID == seedMemberID(6) && members[i].DepartmentID == contract.IDDept3 {
			if !org.ContainsRole(members[i].Roles, permission.RoleOrgAdmin) {
				members[i].Roles = append(members[i].Roles, permission.RoleOrgAdmin)
			}
		}
	}

	for i := range members {
		if members[i].DepartmentID == contract.IDDept6 && members[i].ID != contract.IDMember2 {
			if !org.ContainsRole(members[i].Roles, permission.RoleBudgetApprover) {
				members[i].Roles = append(members[i].Roles, permission.RoleBudgetApprover)
				break
			}
		}
	}

	auditorCount := 0
	for i := range members {
		if members[i].DepartmentID == contract.IDDept8 && members[i].Status == "active" && auditorCount < 3 {
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
