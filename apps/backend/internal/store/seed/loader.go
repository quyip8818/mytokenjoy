package seed

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func Load(cfg config.Config) store.Snapshot {
	members := BuildMembers()
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
	return store.Snapshot{
		Company:         DefaultCompany(cfg),
		OrgIntegration:  orgIntegration,
		SyncLogs:        buildSyncLogs(cfg.DemoToday),
		ImportFailures:  buildImportFailures(),
		OrgNodes:        buildOrgNodes(),
		ModelAllowlist:  buildModelAllowlist(),
		Members:         members,
		Roles:           roles,
		Permissions:     buildPermissions(),
		BudgetGroups:    buildBudgetGroups(),
		BudgetApprovals: buildBudgetApprovals(),
		OverrunPolicy:   buildOverrunPolicy(),
		AlertRules:      buildAlertRules(),
		ProviderKeys:    buildProviderKeys(cfg.DemoToday),
		PlatformKeys:    loadPlatformKeys(),
		Approvals:       buildApprovals(),
		Models:          buildModels(),
		AuditSettings:   buildAuditSettings(),
		OperationLogs:   loadOperationLogs(),
		UsageLedger:     loadUsageLedger(),
	}
}
