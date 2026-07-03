package store

// Store aggregates domain repositories. Naming: "Org" is the organization-management
// domain (members, roles, integration, nodes); OrgNode (Org().Nodes()) is the org-tree
// entity backed by org_nodes. company_id is the tenant boundary.
import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Snapshot struct {
	Company        Company
	OrgIntegration types.OrgIntegration
	SyncLogs       []types.SyncLog
	ImportFailures []types.ImportFailure
	OrgNodes       []types.OrgNode
	ModelAllowlist []ModelAllowlistRow
	Members        []types.Member
	Roles          []types.Role
	Permissions    []types.Permission
	BudgetGroups   []types.BudgetGroup
	OverrunPolicy  types.OverrunPolicyConfig
	AlertRules     []types.AlertRule
	ProviderKeys   []types.ProviderKey
	PlatformKeys   []types.PlatformKey
	Approvals      []types.KeyApproval
	Models         []types.ModelInfo
	AuditSettings  types.AuditSettings
	OperationLogs  []types.OperationLog
	UsageLedger    []types.UsageLedgerEntry
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
	Ledger() LedgerRepository
	Relay() RelayRepository
	SchedulerLock() SchedulerLockRepository
	Usage() UsageRepository
	Notification() NotificationRepository
	WithTx(ctx context.Context, fn func(Store) error) error
}
