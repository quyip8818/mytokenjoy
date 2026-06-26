package store

import "github.com/tokenjoy/backend/internal/domain/types"

type Snapshot struct {
	DataSourceStatus types.DataSourceStatus
	SyncConfig       types.SyncConfig
	SyncLogs         []types.SyncLog
	ImportFailures   []types.ImportFailure
	Departments      []types.Department
	Members          []types.Member
	Roles            []types.Role
	Permissions      []types.Permission
	BudgetTree       []types.BudgetNode
	BudgetGroups     []types.BudgetGroup
	OverrunPolicy    types.OverrunPolicyConfig
	AlertRules       []types.AlertRule
	MemberQuotaPools map[string]types.MemberQuotaPool
	ProviderKeys     []types.ProviderKey
	PlatformKeys     []types.PlatformKey
	Approvals        []types.KeyApproval
	Models           []types.ModelInfo
	RoutingRules     []types.RoutingRule
	ModelUsage       []types.ModelUsage
	TeamUsage        []types.TeamUsage
	AuditSettings    types.AuditSettings
	OperationLogs    []types.OperationLog
	CallLogs         []types.CallLog
}

type Store interface {
	Org() OrgRepository
	Budget() BudgetRepository
	Keys() KeysRepository
	Models() ModelsRepository
	Dashboard() DashboardRepository
	Audit() AuditRepository
}

type OrgRepository interface {
	DataSourceStatus() types.DataSourceStatus
	SetDataSourceStatus(status types.DataSourceStatus)
	ImportFailures() []types.ImportFailure
	SyncConfig() types.SyncConfig
	SetSyncConfig(cfg types.SyncConfig)
	SyncLogs() []types.SyncLog
	Departments() []types.Department
	SetDepartments(departments []types.Department)
	Members() []types.Member
	SetMembers(members []types.Member)
	Roles() []types.Role
	SetRoles(roles []types.Role)
	Permissions() []types.Permission
}

type BudgetRepository interface {
	Tree() []types.BudgetNode
	SetTree(tree []types.BudgetNode)
	Groups() []types.BudgetGroup
	SetGroups(groups []types.BudgetGroup)
	OverrunPolicy() types.OverrunPolicyConfig
	SetOverrunPolicy(policy types.OverrunPolicyConfig)
	AlertRules() []types.AlertRule
	SetAlertRules(rules []types.AlertRule)
	MemberQuotaPools() map[string]types.MemberQuotaPool
	SetMemberQuotaPools(pools map[string]types.MemberQuotaPool)
}

type KeysRepository interface {
	ProviderKeys() []types.ProviderKey
	SetProviderKeys(keys []types.ProviderKey)
	PlatformKeys() []types.PlatformKey
	SetPlatformKeys(keys []types.PlatformKey)
	Approvals() []types.KeyApproval
	SetApprovals(approvals []types.KeyApproval)
}

type ModelsRepository interface {
	Models() []types.ModelInfo
	SetModels(models []types.ModelInfo)
	RoutingRules() []types.RoutingRule
	SetRoutingRules(rules []types.RoutingRule)
}

type DashboardRepository interface {
	ModelUsage() []types.ModelUsage
	TeamUsage() []types.TeamUsage
}

type AuditRepository interface {
	Settings() types.AuditSettings
	SetSettings(settings types.AuditSettings)
	OperationLogs() []types.OperationLog
	CallLogs() []types.CallLog
}
