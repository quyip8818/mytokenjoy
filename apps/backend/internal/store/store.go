package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

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
	Relay() RelayRepository
	WithTx(ctx context.Context, fn func(Store) error) error
}

type OrgRepository interface {
	DataSourceStatus() types.DataSourceStatus
	SetDataSourceStatus(status types.DataSourceStatus) error
	ImportFailures() []types.ImportFailure
	SyncConfig() types.SyncConfig
	SetSyncConfig(cfg types.SyncConfig) error
	SyncLogs() []types.SyncLog
	Departments() []types.Department
	SetDepartments(departments []types.Department) error
	Members() []types.Member
	SetMembers(members []types.Member) error
	Roles() []types.Role
	SetRoles(roles []types.Role) error
	Permissions() []types.Permission
}

type BudgetRepository interface {
	Tree() []types.BudgetNode
	SetTree(tree []types.BudgetNode) error
	Groups() []types.BudgetGroup
	SetGroups(groups []types.BudgetGroup) error
	OverrunPolicy() types.OverrunPolicyConfig
	SetOverrunPolicy(policy types.OverrunPolicyConfig) error
	AlertRules() []types.AlertRule
	SetAlertRules(rules []types.AlertRule) error
	MemberQuotaPools() map[string]types.MemberQuotaPool
	SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error
}

type KeysRepository interface {
	ProviderKeys() []types.ProviderKey
	SetProviderKeys(keys []types.ProviderKey) error
	PlatformKeys() []types.PlatformKey
	SetPlatformKeys(keys []types.PlatformKey) error
	Approvals() []types.KeyApproval
	SetApprovals(approvals []types.KeyApproval) error
}

type ModelsRepository interface {
	Models() []types.ModelInfo
	SetModels(models []types.ModelInfo) error
	RoutingRules() []types.RoutingRule
	SetRoutingRules(rules []types.RoutingRule) error
}

type DashboardRepository interface {
	ModelUsage() []types.ModelUsage
	TeamUsage() []types.TeamUsage
}

type AuditRepository interface {
	Settings() types.AuditSettings
	SetSettings(settings types.AuditSettings) error
	OperationLogs() []types.OperationLog
	CallLogs() []types.CallLog
}
