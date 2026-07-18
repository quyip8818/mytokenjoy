package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildDepartments() []types.Department {
	dept2Parent := contract.IDDept1
	dept3Parent := contract.IDDept2
	dept4Parent := contract.IDDept2
	dept5Parent := contract.IDDept2
	dept6Parent := contract.IDDept1
	dept7Parent := contract.IDDept1
	dept8Parent := contract.IDDept1
	return []types.Department{
		{
			ID: contract.IDDept1, Name: "总公司", ParentID: nil, MemberCount: 41,
			Children: []types.Department{
				{
					ID: contract.IDDept2, Name: "技术部", ParentID: &dept2Parent, MemberCount: 21,
					Children: []types.Department{
						{ID: contract.IDDept3, Name: "后端组", ParentID: &dept3Parent, MemberCount: 8},
						{ID: contract.IDDept4, Name: "前端组", ParentID: &dept4Parent, MemberCount: 7},
						{ID: contract.IDDept5, Name: "测试组", ParentID: &dept5Parent, MemberCount: 6},
					},
				},
				{ID: contract.IDDept6, Name: "产品部", ParentID: &dept6Parent, MemberCount: 6},
				{ID: contract.IDDept7, Name: "市场部", ParentID: &dept7Parent, MemberCount: 6},
				{ID: contract.IDDept8, Name: "行政部", ParentID: &dept8Parent, MemberCount: 7},
			},
		},
	}
}

func buildImportFailures() []types.ImportFailure {
	return []types.ImportFailure{
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000003001"), Name: "李四", EmployeeID: "10087", Reason: "手机号为空"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000003002"), Name: "王五", EmployeeID: "10088", Reason: "部门不存在"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000003003"), Name: "陈静", EmployeeID: "10089", Reason: "邮箱格式错误"},
	}
}

func buildSyncLogs(refDate string) []types.SyncLog {
	return []types.SyncLog{
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004001"), Time: refDate + " 02:00", Type: "scheduled", Result: "success", Detail: "新增 3 人，更新 12 人"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004002"), Time: "2026-06-18 14:00", Type: "manual", Result: "success", Detail: "无变更"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004003"), Time: "2026-06-18 02:00", Type: "scheduled", Result: "partial_failure", Detail: "成功 38 人，失败 3 人"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004004"), Time: "2026-06-17 14:00", Type: "scheduled", Result: "success", Detail: "新增 1 人"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004005"), Time: "2026-06-17 02:00", Type: "scheduled", Result: "success", Detail: "部门结构同步完成"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004006"), Time: "2026-06-16 14:00", Type: "manual", Result: "failure", Detail: "数据源连接超时"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004007"), Time: "2026-06-16 02:00", Type: "scheduled", Result: "failure", Detail: "需软删除 15 人，超过保护阈值 10 人，同步已终止并已通知超管"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004008"), Time: "2026-06-15 14:00", Type: "scheduled", Result: "success", Detail: "新增 2 人"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000004009"), Time: "2026-06-15 02:00", Type: "scheduled", Result: "partial_failure", Detail: "成功 36 人，失败 2 人"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-00000000400a"), Time: "2026-06-14 14:00", Type: "manual", Result: "success", Detail: "无变更"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-00000000400b"), Time: "2026-06-14 02:00", Type: "scheduled", Result: "success", Detail: "全量同步完成"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-00000000400c"), Time: "2026-06-13 02:00", Type: "scheduled", Result: "success", Detail: "初始化同步 41 人"},
	}
}

func buildRoles(members []types.Member) []types.Role {
	roles := []types.Role{
		{ID: contract.IDRole1, Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}, MemberCount: org.CountMembersByRole(members, permission.RoleSuperAdmin)},
		{ID: contract.IDRole2, Name: permission.RoleOrgAdmin, Type: "preset", Permissions: []string{"org:*"}, MemberCount: org.CountMembersByRole(members, permission.RoleOrgAdmin)},
		{ID: contract.IDRole3, Name: permission.RoleMember, Type: "preset", Permissions: []string{"self:*"}, MemberCount: org.CountMembersByRole(members, permission.RoleMember)},
		{ID: contract.IDRole4, Name: permission.RoleAuditor, Type: "preset", Permissions: []string{"audit:read"}, MemberCount: org.CountMembersByRole(members, permission.RoleAuditor)},
		{ID: contract.IDRole5, Name: permission.RoleAPICaller, Type: "preset", Permissions: []string{"api:call"}, MemberCount: org.CountMembersByRole(members, permission.RoleAPICaller)},
		{ID: contract.IDRole6, Name: permission.RoleBudgetApprover, Type: "custom", Permissions: []string{"p-6"}, MemberCount: org.CountMembersByRole(members, permission.RoleBudgetApprover)},
	}
	for i := range roles {
		roles[i].Permissions = mustRoleGrantIDs(roles[i])
	}
	return roles
}

func mustRoleGrantIDs(role types.Role) []string {
	ids, err := permission.RoleGrantIDs(role.Type, role.Name, role.Permissions)
	if err != nil {
		panic(err)
	}
	return ids
}

func buildPermissions() []types.Permission {
	return []types.Permission{
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005001"), Name: "组织架构管理", Group: "组织"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005002"), Name: "成员管理", Group: "组织"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005003"), Name: "角色管理", Group: "组织"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005004"), Name: "数据源配置", Group: "组织"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005005"), Name: "预算分配", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005006"), Name: "预算审批", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005007"), Name: "模型白名单", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005012"), Name: "预算查看", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005013"), Name: "超限策略", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005014"), Name: "模型管理", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005015"), Name: "平台 Key 管理", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005016"), Name: "Provider Key 管理", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005008"), Name: "查看成本看板", Group: "运营"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005009"), Name: "用量分析", Group: "运营"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005010"), Name: "审计日志查看", Group: "运营"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005017"), Name: "我的 Key", Group: "成员"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005018"), Name: "我的审批", Group: "成员"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005011"), Name: "API 调用", Group: "API"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005019"), Name: "组织查看", Group: "组织"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005020"), Name: "Keys 查看", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005021"), Name: "模型查看", Group: "资源管控"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005022"), Name: "账单查看", Group: "运营"},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000005023"), Name: "企业充值", Group: "运营"},
	}
}
