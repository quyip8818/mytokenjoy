package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildBudgetTree() []types.BudgetNode {
	dept2 := "dept-2"
	dept3 := "dept-2"
	dept4 := "dept-2"
	dept5 := "dept-2"
	dept6 := "dept-1"
	dept7 := "dept-1"
	dept8 := "dept-1"
	reserved5000 := 5000.0
	reserved2000 := 2000.0
	reserved1500 := 1500.0
	return []types.BudgetNode{
		{
			ID: "dept-1", Name: "总公司", ParentID: nil,
			Budget: 100000, Consumed: 67500, ReservedPool: &reserved5000, Period: "2026-06",
			Children: []types.BudgetNode{
				{
					ID: "dept-2", Name: "技术部", ParentID: &dept2,
					Budget: 50000, Consumed: 38200, ReservedPool: &reserved2000, Period: "2026-06",
					Children: []types.BudgetNode{
						{ID: IDDept3, Name: "后端组", ParentID: &dept3, Budget: 25000, Consumed: 21000, ReservedPool: &reserved1500, Period: "2026-06"},
						{ID: IDDept4, Name: "前端组", ParentID: &dept4, Budget: 15000, Consumed: 11200, Period: "2026-06"},
						{ID: "dept-5", Name: "测试组", ParentID: &dept5, Budget: 10000, Consumed: 6000, Period: "2026-06"},
					},
				},
				{ID: "dept-6", Name: "产品部", ParentID: &dept6, Budget: 20000, Consumed: 14300, Period: "2026-06"},
				{ID: "dept-7", Name: "市场部", ParentID: &dept7, Budget: 15000, Consumed: 8500, Period: "2026-06"},
				{ID: "dept-8", Name: "行政部", ParentID: &dept8, Budget: 15000, Consumed: 6500, Period: "2026-06"},
			},
		},
	}
}

func buildBudgetGroups() []types.BudgetGroup {
	return []types.BudgetGroup{
		{ID: IDBudgetGroup1, Name: "AI 创新项目组", Budget: 30000, Consumed: 18500, MemberIDs: []string{IDMember1, "m-4", "m-10"}, DepartmentIDs: []string{IDDept3, IDDept4}},
		{ID: "bg-2", Name: "客服 Bot 项目", Budget: 15000, Consumed: 9200, MemberIDs: []string{"m-2"}, DepartmentIDs: []string{"dept-6"}},
		{ID: "bg-3", Name: "市场增长实验", Budget: 12000, Consumed: 6800, MemberIDs: []string{"m-15", "m-20"}, DepartmentIDs: []string{"dept-7"}},
		{ID: IDBudgetGroup4, Name: "内部效率工具", Budget: 8000, Consumed: 4200, MemberIDs: []string{"m-6", "m-8"}, DepartmentIDs: []string{"dept-5"}},
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
		{ID: "alert-1", NodeID: "dept-1", NodeName: "总公司", Thresholds: []int{80, 90, 100}, NotifyRoleIDs: []string{"role-1"}, Enabled: true},
		{ID: "alert-2", NodeID: "dept-2", NodeName: "技术部", Thresholds: []int{80, 90, 100}, NotifyRoleIDs: []string{"role-2"}, Enabled: true},
		{ID: "alert-3", NodeID: IDDept3, NodeName: "后端组", Thresholds: []int{90, 100}, NotifyRoleIDs: []string{"role-2"}, Enabled: true},
		{ID: "alert-4", NodeID: "dept-6", NodeName: "产品部", Thresholds: []int{80, 100}, NotifyRoleIDs: []string{"role-6"}, Enabled: false},
		{ID: "alert-5", NodeID: "dept-4", NodeName: "前端组", Thresholds: []int{80, 90}, NotifyRoleIDs: []string{"role-2"}, Enabled: true},
		{ID: "alert-6", NodeID: "dept-5", NodeName: "测试组", Thresholds: []int{90, 100}, NotifyRoleIDs: []string{"role-2"}, Enabled: true},
		{ID: "alert-7", NodeID: "dept-7", NodeName: "市场部", Thresholds: []int{80, 100}, NotifyRoleIDs: []string{"role-6"}, Enabled: true},
		{ID: "alert-8", NodeID: "dept-8", NodeName: "行政部", Thresholds: []int{100}, NotifyRoleIDs: []string{"role-1"}, Enabled: false},
	}
}
