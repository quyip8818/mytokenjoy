package seed

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func Load(cfg config.Config) store.Snapshot {
	members := BuildMembers()
	roles := buildRoles(members)
	return store.Snapshot{
		DataSourceStatus: types.DataSourceStatus{
			Platform: nil, Connected: false, LastImport: nil, LastImportResult: nil,
		},
		SyncConfig: types.SyncConfig{
			Enabled: false, StartTime: "02:00", FrequencyHours: 12,
			DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
			NotifyPhone: true, NotifyEmail: true, NotifyIm: true,
		},
		SyncLogs:         buildSyncLogs(cfg.DemoToday),
		ImportFailures:   buildImportFailures(),
		Departments:      buildDepartments(),
		Members:          members,
		Roles:            roles,
		Permissions:      buildPermissions(),
		BudgetTree:       buildBudgetTree(),
		BudgetGroups:     buildBudgetGroups(),
		OverrunPolicy:    buildOverrunPolicy(),
		AlertRules:       buildAlertRules(),
		MemberQuotaPools: buildMemberQuotaPools(),
		ProviderKeys:     buildProviderKeys(cfg.DemoToday),
		PlatformKeys:     loadPlatformKeys(),
		Approvals:        buildApprovals(),
		Models:           buildModels(),
		RoutingRules:     buildRoutingRules(),
		ModelUsage:       buildModelUsage(),
		TeamUsage:        buildTeamUsage(),
		AuditSettings:    buildAuditSettings(),
		OperationLogs:    loadOperationLogs(),
		CallLogs:         loadCallLogs(),
	}
}
