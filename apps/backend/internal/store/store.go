package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Snapshot struct {
	Company             Company
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
	AuditSettings       types.AuditSettings
	OperationLogs       []types.OperationLog
	CallLogs            []types.CallLog
	CredentialPlatform  *types.Platform
	EncryptedCredential []byte
}

type Store interface {
	Company() CompanyRepository
	Invite() InviteRepository
	Platform() PlatformRepository
	Billing() BillingRepository
	Org() OrgRepository
	Budget() BudgetRepository
	Keys() KeysRepository
	Models() ModelsRepository
	Audit() AuditRepository
	Relay() RelayRepository
	Credential() CredentialRepository
	SchedulerLock() SchedulerLockRepository
	Usage() UsageRepository
	Notification() NotificationRepository
	WithTx(ctx context.Context, fn func(Store) error) error
}

type OrgRepository interface {
	DataSourceStatus(ctx context.Context) (types.DataSourceStatus, error)
	SetDataSourceStatus(ctx context.Context, status types.DataSourceStatus) error
	ImportFailures(ctx context.Context) ([]types.ImportFailure, error)
	SetImportFailures(ctx context.Context, failures []types.ImportFailure) error
	SyncConfig(ctx context.Context) (types.SyncConfig, error)
	SetSyncConfig(ctx context.Context, cfg types.SyncConfig) error
	SyncLogs(ctx context.Context) ([]types.SyncLog, error)
	AppendSyncLog(ctx context.Context, log types.SyncLog) error
	Departments(ctx context.Context) ([]types.Department, error)
	SetDepartments(ctx context.Context, departments []types.Department) error
	Members(ctx context.Context) ([]types.Member, error)
	SetMembers(ctx context.Context, members []types.Member) error
	SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error
	Roles(ctx context.Context) ([]types.Role, error)
	SetRoles(ctx context.Context, roles []types.Role) error
	Permissions(ctx context.Context) ([]types.Permission, error)
}

type CredentialRepository interface {
	GetCredential(ctx context.Context) (*types.StoredCredential, error)
	SaveCredential(ctx context.Context, platform types.Platform, encrypted []byte) error
	ClearCredential(ctx context.Context) error
}

type SchedulerLockRepository interface {
	TryAcquire(ctx context.Context, lockName, holder string, lease time.Duration) (bool, error)
	Release(ctx context.Context, lockName, holder string) error
}

type BudgetRepository interface {
	Tree(ctx context.Context) ([]types.BudgetNode, error)
	SetTree(ctx context.Context, tree []types.BudgetNode) error
	Groups(ctx context.Context) ([]types.BudgetGroup, error)
	SetGroups(ctx context.Context, groups []types.BudgetGroup) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
	MemberQuotaPools(ctx context.Context) (map[string]types.MemberQuotaPool, error)
	SetMemberQuotaPools(ctx context.Context, pools map[string]types.MemberQuotaPool) error
}

type KeysRepository interface {
	ProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error
	PlatformKeys(ctx context.Context) ([]types.PlatformKey, error)
	SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error
	Approvals(ctx context.Context) ([]types.KeyApproval, error)
	SetApprovals(ctx context.Context, approvals []types.KeyApproval) error
}

type ModelsRepository interface {
	Models(ctx context.Context) ([]types.ModelInfo, error)
	SetModels(ctx context.Context, models []types.ModelInfo) error
	RoutingRules(ctx context.Context) ([]types.RoutingRule, error)
	SetRoutingRules(ctx context.Context, rules []types.RoutingRule) error
}

type AuditRepository interface {
	Settings(ctx context.Context) (types.AuditSettings, error)
	SetSettings(ctx context.Context, settings types.AuditSettings) error
	OperationLogs(ctx context.Context) ([]types.OperationLog, error)
	AppendOperationLog(ctx context.Context, log types.OperationLog) error
	CallLogs(ctx context.Context) ([]types.CallLog, error)
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
