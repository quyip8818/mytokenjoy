package store

import (
	"context"

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
