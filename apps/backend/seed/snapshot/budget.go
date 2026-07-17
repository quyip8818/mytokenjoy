package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/filler"
)

func buildBudgetTree() []types.BudgetNode {
	dept2 := contract.IDDept1
	dept3 := contract.IDDept2
	dept4 := contract.IDDept2
	dept5 := contract.IDDept2
	dept6 := contract.IDDept1
	dept7 := contract.IDDept1
	dept8 := contract.IDDept1
	reserved15000 := seedPoints(15000)
	reserved2000 := seedPoints(2000)
	reserved1500 := seedPoints(1500)
	techConsumed := contract.DemoLeafDeptConsumed[contract.IDDept3] +
		contract.DemoLeafDeptConsumed[contract.IDDept4] +
		contract.DemoLeafDeptConsumed[contract.IDDept5]
	return []types.BudgetNode{
		{
			ID: contract.IDDept1, Name: "总公司", ParentID: nil,
			Budget: seedPoints(120000), Consumed: contract.DemoRootConsumed(), ReservedPool: &reserved15000, Period: pkgbudget.PeriodMonthly,
			Children: []types.BudgetNode{
				{
					ID: contract.IDDept2, Name: "技术部", ParentID: &dept2,
					Budget: seedPoints(50000), Consumed: techConsumed, ReservedPool: &reserved2000, Period: pkgbudget.PeriodMonthly,
					Children: []types.BudgetNode{
						{ID: contract.IDDept3, Name: "后端组", ParentID: &dept3, Budget: seedPoints(20000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept3], ReservedPool: &reserved1500, Period: pkgbudget.PeriodMonthly},
						{ID: contract.IDDept4, Name: "前端组", ParentID: &dept4, Budget: seedPoints(15000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept4], Period: pkgbudget.PeriodMonthly},
						{ID: contract.IDDept5, Name: "测试组", ParentID: &dept5, Budget: seedPoints(10000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept5], Period: pkgbudget.PeriodMonthly},
					},
				},
				{ID: contract.IDDept6, Name: "产品部", ParentID: &dept6, Budget: seedPoints(20000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept6], Period: pkgbudget.PeriodMonthly},
				{ID: contract.IDDept7, Name: "市场部", ParentID: &dept7, Budget: seedPoints(15000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept7], Period: pkgbudget.PeriodMonthly},
				{ID: contract.IDDept8, Name: "行政部", ParentID: &dept8, Budget: seedPoints(16000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept8], Period: pkgbudget.PeriodMonthly},
			},
		},
	}
}

func buildProjects() []types.Project {
	return []types.Project{
		{
			ID: contract.IDProject1, Name: "AI 创新项目组", Budget: seedPoints(15000),
			Consumed:  contract.DemoProjectConsumed[contract.IDProject1],
			MemberIDs: []uuid.UUID{contract.IDMember1, contract.IDMember4, contract.IDMember6},
			MemberBudgets: map[uuid.UUID]float64{
				contract.IDMember1: seedPoints(6000),
				contract.IDMember4: seedPoints(5000),
				contract.IDMember6: seedPoints(3000),
			},
			OwnerDepartmentID: contract.IDDept3,
		},
		{
			ID: contract.IDProject4, Name: "内部效率工具", Budget: seedPoints(8000),
			Consumed:  contract.DemoProjectConsumed[contract.IDProject4],
			MemberIDs: []uuid.UUID{contract.IDMember15, contract.IDMember16},
			MemberBudgets: map[uuid.UUID]float64{
				contract.IDMember15: seedPoints(4000),
				contract.IDMember16: seedPoints(4000),
			},
			OwnerDepartmentID: contract.IDDept5,
		},
	}
}

func buildOverrunPolicy() types.OverrunPolicyConfig {
	return types.OverrunPolicyConfig{
		Thresholds: []int{80, 90}, NotifyEmail: true, NotifyPhone: true, NotifyIm: true,
		BlockMessage: "额度已用尽，请联系管理员申请追加",
	}
}

func buildAlertRules() []types.AlertRule {
	return []types.AlertRule{
		{ID: contract.IDAlertRule1, NodeID: contract.IDDept1, NodeName: "总公司", Thresholds: []int{80, 90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole1}, Enabled: true},
		{ID: contract.IDAlertRule2, NodeID: contract.IDDept2, NodeName: "技术部", Thresholds: []int{80, 90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule3, NodeID: contract.IDDept3, NodeName: "后端组", Thresholds: []int{90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule4, NodeID: contract.IDDept6, NodeName: "产品部", Thresholds: []int{80, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole6}, Enabled: false},
		{ID: contract.IDAlertRule5, NodeID: contract.IDDept4, NodeName: "前端组", Thresholds: []int{80, 90}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule6, NodeID: contract.IDDept5, NodeName: "测试组", Thresholds: []int{90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule7, NodeID: contract.IDDept7, NodeName: "市场部", Thresholds: []int{80, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole6}, Enabled: true},
		{ID: contract.IDAlertRule8, NodeID: contract.IDDept8, NodeName: "行政部", Thresholds: []int{100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole1}, Enabled: false},
	}
}

func buildBudgetApprovals() []types.BudgetApproval {
	resolved1 := "2026-06-20 11:30"
	resolved2 := "2026-06-15 16:45"
	resolved3 := "2026-06-25 17:30"
	productMember := firstMemberInDepartment(contract.IDDept6)
	return []types.BudgetApproval{
		{
			ID: contract.IDBudgetApproval1, ApplicantID: contract.IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 500, Reason: "本月额度用尽，需完成搜索优化任务",
			Status: "pending", CreatedAt: "2026-06-28 14:30",
		},
		{
			ID: contract.IDBudgetApproval2, ApplicantID: contract.IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 300, Reason: "RAG 管道调试需额外调用",
			Status: "approved", CreatedAt: "2026-06-20 09:00", ResolvedAt: &resolved1,
		},
		{
			ID: contract.IDBudgetApproval3, ApplicantID: contract.IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 200, Reason: "紧急修复线上搜索问题",
			Status: "approved", CreatedAt: "2026-06-15 16:00", ResolvedAt: &resolved2,
		},
		{
			ID: contract.IDBudgetAppr4, ApplicantID: contract.IDMember4, ApplicantName: "赵六", DepartmentName: "后端组",
			Amount: 300, Reason: "调试 RAG 管道需额外调用",
			Status: "pending", CreatedAt: "2026-06-29 09:15",
		},
		{
			ID: contract.IDBudgetAppr5, ApplicantID: productMember.ID, ApplicantName: productMember.Name, DepartmentName: productMember.DepartmentName,
			Amount: 200, Reason: "产品文档生成",
			Status: "approved", CreatedAt: "2026-06-25 16:00", ResolvedAt: &resolved3,
		},
	}
}

func firstMemberInDepartment(departmentID uuid.UUID) types.Member {
	for _, member := range filler.BuildMembers() {
		if member.DepartmentID == departmentID {
			return member
		}
	}
	return types.Member{ID: contract.IDMemberFallback, Name: "产品成员", DepartmentName: "产品部"}
}
