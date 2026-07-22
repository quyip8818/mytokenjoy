package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildBudgetTree() []types.BudgetNode {
	dept2 := contract.IDDept1
	dept3 := contract.IDDept2
	dept4 := contract.IDDept2
	dept5 := contract.IDDept2
	dept6 := contract.IDDept1
	dept7 := contract.IDDept1
	dept8 := contract.IDDept1
	reserved15000 := seedQuota(15000)
	reserved2000 := seedQuota(2000)
	reserved1500 := seedQuota(1500)
	techConsumed := contract.DemoLeafDeptConsumed[contract.IDDept3] +
		contract.DemoLeafDeptConsumed[contract.IDDept4] +
		contract.DemoLeafDeptConsumed[contract.IDDept5]
	return []types.BudgetNode{
		{
			ID: contract.IDDept1, Name: "总公司", ParentID: nil,
			Budget: seedQuota(120000), Consumed: contract.DemoRootConsumed(), ReservedPool: &reserved15000, Period: pkgbudget.PeriodMonthly,
			Children: []types.BudgetNode{
				{
					ID: contract.IDDept2, Name: "技术部", ParentID: &dept2,
					Budget: seedQuota(50000), Consumed: techConsumed, ReservedPool: &reserved2000, Period: pkgbudget.PeriodMonthly,
					Children: []types.BudgetNode{
						{ID: contract.IDDept3, Name: "后端组", ParentID: &dept3, Budget: seedQuota(20000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept3], ReservedPool: &reserved1500, Period: pkgbudget.PeriodMonthly},
						{ID: contract.IDDept4, Name: "前端组", ParentID: &dept4, Budget: seedQuota(15000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept4], Period: pkgbudget.PeriodMonthly},
						{ID: contract.IDDept5, Name: "测试组", ParentID: &dept5, Budget: seedQuota(10000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept5], Period: pkgbudget.PeriodMonthly},
					},
				},
				{ID: contract.IDDept6, Name: "产品部", ParentID: &dept6, Budget: seedQuota(20000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept6], Period: pkgbudget.PeriodMonthly},
				{ID: contract.IDDept7, Name: "市场部", ParentID: &dept7, Budget: seedQuota(15000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept7], Period: pkgbudget.PeriodMonthly},
				{ID: contract.IDDept8, Name: "行政部", ParentID: &dept8, Budget: seedQuota(16000), Consumed: contract.DemoLeafDeptConsumed[contract.IDDept8], Period: pkgbudget.PeriodMonthly},
			},
		},
	}
}

func buildProjects() []types.Project {
	return []types.Project{
		{
			ID: contract.IDProject1, Name: "AI 创新项目组", Budget: seedQuota(15000),
			Consumed:  contract.DemoProjectConsumed[contract.IDProject1],
			MemberIDs: []uuid.UUID{contract.IDMember1, contract.IDMember4, contract.IDMember6},
			MemberBudgets: map[uuid.UUID]int64{
				contract.IDMember1: seedQuota(6000),
				contract.IDMember4: seedQuota(5000),
				contract.IDMember6: seedQuota(3000),
			},
			OwnerDepartmentID: contract.IDDept3,
		},
		{
			ID: contract.IDProject4, Name: "内部效率工具", Budget: seedQuota(8000),
			Consumed:  contract.DemoProjectConsumed[contract.IDProject4],
			MemberIDs: []uuid.UUID{contract.IDMember15, contract.IDMember16},
			MemberBudgets: map[uuid.UUID]int64{
				contract.IDMember15: seedQuota(4000),
				contract.IDMember16: seedQuota(4000),
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
		{ID: contract.IDAlertRule4, NodeID: contract.IDDept6, NodeName: "产品部", Thresholds: []int{80, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRoleBudgetApprover}, Enabled: false},
		{ID: contract.IDAlertRule5, NodeID: contract.IDDept4, NodeName: "前端组", Thresholds: []int{80, 90}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule6, NodeID: contract.IDDept5, NodeName: "测试组", Thresholds: []int{90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
		{ID: contract.IDAlertRule7, NodeID: contract.IDDept7, NodeName: "市场部", Thresholds: []int{80, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRoleBudgetApprover}, Enabled: true},
		{ID: contract.IDAlertRule8, NodeID: contract.IDDept8, NodeName: "行政部", Thresholds: []int{100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole1}, Enabled: false},
	}
}
