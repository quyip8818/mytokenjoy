package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildBudgetApprovals() []types.BudgetApproval {
	resolved1 := "2026-06-20 11:30"
	resolved2 := "2026-06-15 16:45"
	resolved3 := "2026-06-25 17:30"
	productMember := firstMemberInDepartment("dept-6")
	return []types.BudgetApproval{
		{
			ID: "appr-1", ApplicantID: IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 500, Reason: "本月额度用尽，需完成搜索优化任务",
			Status: "pending", CreatedAt: "2026-06-28 14:30",
		},
		{
			ID: "appr-1b", ApplicantID: IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 300, Reason: "RAG 管道调试需额外调用",
			Status: "approved", CreatedAt: "2026-06-20 09:00", ResolvedAt: &resolved1,
		},
		{
			ID: "appr-1c", ApplicantID: IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 200, Reason: "紧急修复线上搜索问题",
			Status: "approved", CreatedAt: "2026-06-15 16:00", ResolvedAt: &resolved2,
		},
		{
			ID: "appr-2", ApplicantID: "m-4", ApplicantName: "赵六", DepartmentName: "后端组",
			Amount: 300, Reason: "调试 RAG 管道需额外调用",
			Status: "pending", CreatedAt: "2026-06-29 09:15",
		},
		{
			ID: "appr-3", ApplicantID: productMember.ID, ApplicantName: productMember.Name, DepartmentName: productMember.DepartmentName,
			Amount: 200, Reason: "产品文档生成",
			Status: "approved", CreatedAt: "2026-06-25 16:00", ResolvedAt: &resolved3,
		},
	}
}

func firstMemberInDepartment(departmentID string) types.Member {
	for _, member := range BuildMembers() {
		if member.DepartmentID == departmentID {
			return member
		}
	}
	return types.Member{ID: "m-45", Name: "产品成员", DepartmentName: "产品部"}
}
