package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Snapshot struct {
	DataSourceStatus    types.DataSourceStatus
	SyncConfig          types.SyncConfig
	SyncLogs            []types.SyncLog
	ImportFailures      []types.ImportFailure
	Departments         []types.Department
	Members             []types.Member
	Roles               []types.Role
	Permissions         []types.Permission
	BudgetTree          []types.BudgetNode
	BudgetGroups        []types.BudgetGroup
	OverrunPolicy       types.OverrunPolicyConfig
	AlertRules          []types.AlertRule
	MemberQuotaPools    map[string]types.MemberQuotaPool
	ProviderKeys        []types.ProviderKey
	PlatformKeys        []types.PlatformKey
	Approvals           []types.KeyApproval
	Models              []types.ModelInfo
	RoutingRules        []types.RoutingRule
	ModelUsage          []types.ModelUsage
	TeamUsage           []types.TeamUsage
	AuditSettings       types.AuditSettings
	OperationLogs       []types.OperationLog
	CallLogs            []types.CallLog
	CredentialPlatform  *types.Platform
	EncryptedCredential []byte
}

type Store interface {
	Org() OrgRepository
	Budget() BudgetRepository
	Keys() KeysRepository
	Models() ModelsRepository
	Dashboard() DashboardRepository
	Audit() AuditRepository
	Relay() RelayRepository
	Credential() CredentialRepository
	SchedulerLock() SchedulerLockRepository
	Usage() UsageRepository
	Notification() NotificationRepository
	WithTx(ctx context.Context, fn func(Store) error) error
}

type OrgRepository interface {
	DataSourceStatus() types.DataSourceStatus
	SetDataSourceStatus(status types.DataSourceStatus) error
	ImportFailures() []types.ImportFailure
	SetImportFailures(failures []types.ImportFailure) error
	SyncConfig() types.SyncConfig
	SetSyncConfig(cfg types.SyncConfig) error
	SyncLogs() []types.SyncLog
	AppendSyncLog(log types.SyncLog) error
	Departments() []types.Department
	SetDepartments(departments []types.Department) error
	Members() []types.Member
	SetMembers(members []types.Member) error
	Roles() []types.Role
	SetRoles(roles []types.Role) error
	Permissions() []types.Permission
}

type CredentialRepository interface {
	GetCredential() (*types.StoredCredential, error)
	SaveCredential(platform types.Platform, encrypted []byte) error
	ClearCredential() error
}

type SchedulerLockRepository interface {
	TryAcquire(ctx context.Context, lockName, holder string, lease time.Duration) (bool, error)
	Release(ctx context.Context, lockName, holder string) error
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

type UsageRepository interface {
	UpsertBucket(ctx context.Context, row types.UsageBucketRow) error
	QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
	QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error)
	QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error)
}

type NotificationRepository interface {
	Append(ctx context.Context, entry types.NotificationLogEntry) error
}
