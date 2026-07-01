package store

func CloneSnapshot(snapshot Snapshot) Snapshot {
	return Snapshot{
		Company:             snapshot.Company,
		DataSourceStatus:    snapshot.DataSourceStatus,
		SyncConfig:          snapshot.SyncConfig,
		SyncLogs:            CloneSyncLogs(snapshot.SyncLogs),
		ImportFailures:      CloneImportFailures(snapshot.ImportFailures),
		Departments:         CloneDepartments(snapshot.Departments),
		Members:             CloneMembers(snapshot.Members),
		Roles:               CloneRoles(snapshot.Roles),
		Permissions:         ClonePermissions(snapshot.Permissions),
		BudgetTree:          CloneBudgetTree(snapshot.BudgetTree),
		BudgetGroups:        CloneBudgetGroups(snapshot.BudgetGroups),
		OverrunPolicy:       snapshot.OverrunPolicy,
		AlertRules:          CloneAlertRules(snapshot.AlertRules),
		MemberQuotaPools:    CloneMemberQuotaPools(snapshot.MemberQuotaPools),
		ProviderKeys:        CloneProviderKeys(snapshot.ProviderKeys),
		PlatformKeys:        ClonePlatformKeys(snapshot.PlatformKeys),
		Approvals:           CloneApprovals(snapshot.Approvals),
		Models:              CloneModels(snapshot.Models),
		RoutingRules:        CloneRoutingRules(snapshot.RoutingRules),
		AuditSettings:       snapshot.AuditSettings,
		OperationLogs:       CloneOperationLogs(snapshot.OperationLogs),
		CallLogs:            CloneCallLogs(snapshot.CallLogs),
		CredentialPlatform:  snapshot.CredentialPlatform,
		EncryptedCredential: append([]byte(nil), snapshot.EncryptedCredential...),
	}
}
