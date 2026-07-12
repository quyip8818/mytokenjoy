package store

// Store aggregates domain repositories. Naming: "Org" is the organization-management
// domain (members, roles, integration, nodes); OrgNode (Org().Nodes()) is the org-tree
// entity backed by org_nodes. company_id is the tenant boundary.
import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Snapshot struct {
	Company         Company
	OrgIntegration  types.OrgIntegration
	SyncLogs        []types.SyncLog
	ImportFailures  []types.ImportFailure
	OrgNodes        []types.OrgNode
	ModelAllowlist  []ModelAllowlistRow
	Members         []types.Member
	Roles           []types.Role
	Permissions     []types.Permission
	Projects        []types.Project
	OverrunPolicy   types.OverrunPolicyConfig
	AlertRules      []types.AlertRule
	BudgetApprovals []types.BudgetApproval
	ProviderKeys    []types.ProviderKey
	PlatformKeys    []types.PlatformKey
	Approvals       []types.KeyApproval
	Models          []types.ModelInfo
	AuditSettings   types.AuditSettings
	OperationLogs   []types.OperationLog
	UsageLedger     []types.UsageLedgerEntry
	// SeedAt is the business-clock instant used for open-budget snapshot period keys.
	SeedAt time.Time
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
	PlatformKeyMappings() PlatformKeyMappingRepository
	BudgetConsumed() BudgetConsumedRepository
	BudgetProjectionProgress() ProjectionProgressRepository
	DashboardProjectionProgress() ProjectionProgressRepository
	GatewayPrecheck() GatewayPrecheckRepository
	GatewaySoftSummaries() GatewaySoftSummaryRepository
	SchedulerLock() SchedulerLockRepository
	Usage() UsageRepository
	Notification() NotificationRepository
	Logs() LogStore
	WithTx(ctx context.Context, fn func(Store) error) error
}

// ConsumptionWriter is the minimal store surface for ingest transactions:
// ledger insert and side-effect enqueue.
type ConsumptionWriter interface {
	Ledger() LedgerRepository
	Usage() UsageRepository
	Budget() BudgetRepository
	Org() OrgRepository
	Keys() KeysRepository
	PlatformKeyMappings() PlatformKeyMappingRepository
}
