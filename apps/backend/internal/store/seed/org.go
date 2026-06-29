package seed

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func buildDepartments() []types.Department {
	dept2Parent := "dept-1"
	dept3Parent := "dept-2"
	dept4Parent := "dept-2"
	dept5Parent := "dept-2"
	dept6Parent := "dept-1"
	dept7Parent := "dept-1"
	dept8Parent := "dept-1"
	return []types.Department{
		{
			ID: "dept-1", Name: "总公司", ParentID: nil, MemberCount: 128,
			Children: []types.Department{
				{
					ID: "dept-2", Name: "技术部", ParentID: &dept2Parent, MemberCount: 45,
					Children: []types.Department{
						{ID: IDDept3, Name: "后端组", ParentID: &dept3Parent, MemberCount: 20},
						{ID: IDDept4, Name: "前端组", ParentID: &dept4Parent, MemberCount: 15},
						{ID: "dept-5", Name: "测试组", ParentID: &dept5Parent, MemberCount: 10},
					},
				},
				{ID: "dept-6", Name: "产品部", ParentID: &dept6Parent, MemberCount: 25},
				{ID: "dept-7", Name: "市场部", ParentID: &dept7Parent, MemberCount: 30},
				{ID: "dept-8", Name: "行政部", ParentID: &dept8Parent, MemberCount: 28},
			},
		},
	}
}

func buildImportFailures() []types.ImportFailure {
	return []types.ImportFailure{
		{ID: "f-1", Name: "李四", EmployeeID: "10087", Reason: "手机号为空"},
		{ID: "f-2", Name: "王五", EmployeeID: "10088", Reason: "部门不存在"},
		{ID: "f-3", Name: "陈静", EmployeeID: "10089", Reason: "邮箱格式错误"},
	}
}

func buildSyncLogs(demoToday string) []types.SyncLog {
	return []types.SyncLog{
		{ID: "sync-1", Time: demoToday + " 02:00", Type: "scheduled", Result: "success", Detail: "新增 3 人，更新 12 人"},
		{ID: "sync-2", Time: "2026-06-18 14:00", Type: "manual", Result: "success", Detail: "无变更"},
		{ID: "sync-3", Time: "2026-06-18 02:00", Type: "scheduled", Result: "partial_failure", Detail: "成功 125 人，失败 3 人"},
		{ID: "sync-4", Time: "2026-06-17 14:00", Type: "scheduled", Result: "success", Detail: "新增 1 人"},
		{ID: "sync-5", Time: "2026-06-17 02:00", Type: "scheduled", Result: "success", Detail: "部门结构同步完成"},
		{ID: "sync-6", Time: "2026-06-16 14:00", Type: "manual", Result: "failure", Detail: "数据源连接超时"},
		{ID: "sync-7", Time: "2026-06-16 02:00", Type: "scheduled", Result: "failure", Detail: "需软删除 15 人，超过保护阈值 10 人，同步已终止并已通知超管"},
		{ID: "sync-8", Time: "2026-06-15 14:00", Type: "scheduled", Result: "success", Detail: "新增 2 人"},
		{ID: "sync-9", Time: "2026-06-15 02:00", Type: "scheduled", Result: "partial_failure", Detail: "成功 118 人，失败 2 人"},
		{ID: "sync-10", Time: "2026-06-14 14:00", Type: "manual", Result: "success", Detail: "无变更"},
		{ID: "sync-11", Time: "2026-06-14 02:00", Type: "scheduled", Result: "success", Detail: "全量同步完成"},
		{ID: "sync-12", Time: "2026-06-13 02:00", Type: "scheduled", Result: "success", Detail: "初始化同步 128 人"},
	}
}

func buildRoles(members []types.Member) []types.Role {
	return []types.Role{
		{ID: "role-1", Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}, MemberCount: org.CountMembersByRole(members, permission.RoleSuperAdmin)},
		{ID: "role-2", Name: permission.RoleOrgAdmin, Type: "preset", Permissions: []string{"org:*"}, MemberCount: org.CountMembersByRole(members, permission.RoleOrgAdmin)},
		{ID: "role-3", Name: permission.RoleMember, Type: "preset", Permissions: []string{"self:*"}, MemberCount: org.CountMembersByRole(members, permission.RoleMember)},
		{ID: "role-4", Name: permission.RoleAuditor, Type: "preset", Permissions: []string{"audit:read"}, MemberCount: org.CountMembersByRole(members, permission.RoleAuditor)},
		{ID: "role-5", Name: permission.RoleAPICaller, Type: "preset", Permissions: []string{"api:call"}, MemberCount: org.CountMembersByRole(members, permission.RoleAPICaller)},
		{ID: "role-6", Name: permission.RoleBudgetApprover, Type: "custom", Permissions: []string{"p-6"}, MemberCount: org.CountMembersByRole(members, permission.RoleBudgetApprover)},
	}
}

func buildPermissions() []types.Permission {
	return []types.Permission{
		{ID: "p-1", Name: "组织架构管理", Group: "组织"},
		{ID: "p-2", Name: "成员管理", Group: "组织"},
		{ID: "p-3", Name: "角色管理", Group: "组织"},
		{ID: "p-4", Name: "数据源配置", Group: "组织"},
		{ID: "p-5", Name: "预算分配", Group: "资源管控"},
		{ID: "p-6", Name: "预算审批", Group: "资源管控"},
		{ID: "p-7", Name: "模型白名单", Group: "资源管控"},
		{ID: "p-12", Name: "预算查看", Group: "资源管控"},
		{ID: "p-13", Name: "超限策略", Group: "资源管控"},
		{ID: "p-14", Name: "模型管理", Group: "资源管控"},
		{ID: "p-15", Name: "平台 Key 管理", Group: "资源管控"},
		{ID: "p-16", Name: "Provider Key 管理", Group: "资源管控"},
		{ID: "p-8", Name: "查看成本看板", Group: "运营"},
		{ID: "p-9", Name: "用量分析", Group: "运营"},
		{ID: "p-10", Name: "审计日志查看", Group: "运营"},
		{ID: "p-17", Name: "我的 Key", Group: "成员"},
		{ID: "p-18", Name: "我的审批", Group: "成员"},
		{ID: "p-11", Name: "API 调用", Group: "API"},
		{ID: "p-19", Name: "组织查看", Group: "组织"},
		{ID: "p-20", Name: "Keys 查看", Group: "资源管控"},
		{ID: "p-21", Name: "模型查看", Group: "资源管控"},
	}
}
