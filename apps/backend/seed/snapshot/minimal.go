package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/filler"
)

func BuildMinimal(cfg config.Config) store.Snapshot {
	members := filler.BuildAnchorMembers()
	roles := buildRoles(members)
	orgIntegration := types.OrgIntegrationFromStatusAndConfig(
		types.DataSourceStatus{Platform: nil, Connected: false, LastImport: nil, LastImportResult: nil},
		types.SyncConfig{
			Enabled: false, StartTime: "02:00", FrequencyHours: 12,
			DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
			NotifyPhone: true, NotifyEmail: true, NotifyIm: true,
		},
	)
	orgIntegration.FieldMappings = buildDefaultFieldMappings()
	platformKeys := minimalPlatformKeys()
	ref := cfg.SeedReferenceDate()
	return store.Snapshot{
		Company:        defaultCompany(cfg),
		OrgIntegration: orgIntegration,
		SyncLogs: []types.SyncLog{
			{ID: contract.IDSyncLog1, Time: ref + " 02:00", Type: "scheduled", Result: "success", Detail: "初始化同步 8 人"},
		},
		ImportFailures:  nil,
		OrgNodes:        buildMinimalOrgNodes(),
		ModelAllowlist:  minimalModelAllowlist(platformKeys),
		Members:         members,
		Roles:           roles,
		Permissions:     buildPermissions(),
		Projects:        minimalProjects(),
		BudgetApprovals: minimalBudgetApprovals(),
		OverrunPolicy:   buildOverrunPolicy(),
		AlertRules:      minimalAlertRules(),
		ProviderKeys:    buildProviderKeys()[:3],
		PlatformKeys:    platformKeys,
		Approvals:       buildApprovals()[:1],
		Models:          buildModels(),
		AuditSettings:   buildAuditSettings(),
		OperationLogs:   loadOperationLogs()[:1],
		UsageLedger:     nil,
		SeedAt:          clock.NowUTC(cfg.Clock()),
	}
}

func buildMinimalDepartments() []types.Department {
	dept2Parent := contract.IDDept1
	dept3Parent := contract.IDDept2
	dept4Parent := contract.IDDept2
	dept5Parent := contract.IDDept2
	dept8Parent := contract.IDDept1
	return []types.Department{
		{
			ID: contract.IDDept1, Name: "总公司", ParentID: nil, MemberCount: 8,
			Children: []types.Department{
				{
					ID: contract.IDDept2, Name: "技术部", ParentID: &dept2Parent, MemberCount: 6,
					Children: []types.Department{
						{ID: contract.IDDept3, Name: "后端组", ParentID: &dept3Parent, MemberCount: 4},
						{ID: contract.IDDept4, Name: "前端组", ParentID: &dept4Parent, MemberCount: 2},
						{ID: contract.IDDept5, Name: "测试组", ParentID: &dept5Parent, MemberCount: 0},
					},
				},
				{ID: contract.IDDept8, Name: "行政部", ParentID: &dept8Parent, MemberCount: 1},
			},
		},
	}
}

func buildMinimalOrgNodes() []types.OrgNode {
	return assembleOrgNodes(buildMinimalDepartments())
}

func minimalPlatformKeys() []types.PlatformKey {
	for _, key := range loadPlatformKeys() {
		if key.ID == contract.IDPlatformKey1 {
			return []types.PlatformKey{key}
		}
	}
	return nil
}

func minimalModelAllowlist(keys []types.PlatformKey) []store.ModelAllowlistRow {
	rows := make([]store.ModelAllowlistRow, 0)
	for nodeID, cfg := range orgNodeRoutingByID() {
		if nodeID != contract.IDDept3 && nodeID != contract.IDDept1 {
			continue
		}
		for _, modelID := range cfg.allowedModelIDs {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerOrgNode,
				OwnerID:   nodeID,
				ModelID:   modelID,
			})
		}
	}
	for _, key := range keys {
		for _, modelID := range key.ModelWhitelist {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerPlatformKey,
				OwnerID:   key.ID,
				ModelID:   modelID,
			})
		}
	}
	return rows
}

func minimalProjects() []types.Project {
	return []types.Project{
		{
			ID: contract.IDProject1, Name: "AI 创新项目组", Budget: 30000, Consumed: 18500,
			MemberIDs: []uuid.UUID{contract.IDMember1, contract.IDMember4}, OwnerDepartmentID: contract.IDDept3,
		},
	}
}

func minimalBudgetApprovals() []types.BudgetApproval {
	return []types.BudgetApproval{
		{
			ID: contract.IDBudgetApproval1, ApplicantID: contract.IDMember1, ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 500, Reason: "本月额度用尽，需完成搜索优化任务",
			Status: "pending", CreatedAt: "2026-06-28 14:30",
		},
	}
}

func minimalAlertRules() []types.AlertRule {
	return []types.AlertRule{
		{ID: contract.IDAlertRule1, NodeID: contract.IDDept1, NodeName: "总公司", Thresholds: []int{80, 90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole1}, Enabled: true},
		{ID: contract.IDAlertRule3, NodeID: contract.IDDept3, NodeName: "后端组", Thresholds: []int{90, 100}, NotifyRoleIDs: []uuid.UUID{contract.IDRole2}, Enabled: true},
	}
}
