package store

import "github.com/tokenjoy/backend/internal/domain/types"

func CloneOrgIntegration(integration types.OrgIntegration) types.OrgIntegration {
	cloned := integration
	if integration.EncryptedCredential != nil {
		cloned.EncryptedCredential = append([]byte(nil), integration.EncryptedCredential...)
	}
	if len(integration.FieldMappings) > 0 {
		cloned.FieldMappings = make([]types.FieldMapping, len(integration.FieldMappings))
		copy(cloned.FieldMappings, integration.FieldMappings)
	}
	return cloned
}

func CloneSnapshot(snapshot Snapshot) Snapshot {
	return Snapshot{
		Company:         snapshot.Company,
		OrgIntegration:  CloneOrgIntegration(snapshot.OrgIntegration),
		SyncLogs:        CloneSyncLogs(snapshot.SyncLogs),
		ImportFailures:  CloneImportFailures(snapshot.ImportFailures),
		OrgNodes:        CloneOrgNodes(snapshot.OrgNodes),
		ModelAllowlist:  CloneModelAllowlist(snapshot.ModelAllowlist),
		Members:         CloneMembers(snapshot.Members),
		Roles:           CloneRoles(snapshot.Roles),
		Permissions:     ClonePermissions(snapshot.Permissions),
		BudgetGroups:    CloneBudgetGroups(snapshot.BudgetGroups),
		OverrunPolicy:   snapshot.OverrunPolicy,
		AlertRules:      CloneAlertRules(snapshot.AlertRules),
		BudgetApprovals: CloneBudgetApprovals(snapshot.BudgetApprovals),
		ProviderKeys:    CloneProviderKeys(snapshot.ProviderKeys),
		PlatformKeys:    ClonePlatformKeys(snapshot.PlatformKeys),
		Approvals:       CloneApprovals(snapshot.Approvals),
		Models:          CloneModels(snapshot.Models),
		AuditSettings:   snapshot.AuditSettings,
		OperationLogs:   CloneOperationLogs(snapshot.OperationLogs),
		UsageLedger:     CloneUsageLedger(snapshot.UsageLedger),
	}
}
