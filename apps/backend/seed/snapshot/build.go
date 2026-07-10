package snapshot

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/filler"
)

func Build(cfg config.Config) store.Snapshot {
	members := filler.BuildMembers()
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
	ref := cfg.SeedReferenceDate()
	return store.Snapshot{
		Company:         defaultCompany(cfg),
		OrgIntegration:  orgIntegration,
		SyncLogs:        buildSyncLogs(ref),
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
		ProviderKeys:    buildProviderKeys(ref),
		PlatformKeys:    loadPlatformKeys(),
		Approvals:       buildApprovals(),
		Models:          buildModels(),
		AuditSettings:   buildAuditSettings(),
		OperationLogs:   loadOperationLogs(),
		UsageLedger:     loadUsageLedger(),
		SeedAt:          clock.NowUTC(cfg.Clock()),
	}
}
