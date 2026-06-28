package store

import (
	"sync"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Memory struct {
	mu               sync.RWMutex
	data             Snapshot
	relayRepo        *memoryRelayRepo
	schedulerLocks   map[string]schedulerLockEntry
	usageBuckets     map[string]types.UsageBucketRow
	notificationLogs []types.NotificationLogEntry
}

func NewMemory(snapshot Snapshot) *Memory {
	m := &Memory{data: deepCopySnapshot(snapshot)}
	m.relayRepo = newMemoryRelayRepo(m)
	return m
}

func (m *Memory) Org() OrgRepository             { return &memoryOrgRepo{store: m} }
func (m *Memory) Budget() BudgetRepository       { return &memoryBudgetRepo{store: m} }
func (m *Memory) Keys() KeysRepository           { return &memoryKeysRepo{store: m} }
func (m *Memory) Models() ModelsRepository       { return &memoryModelsRepo{store: m} }
func (m *Memory) Dashboard() DashboardRepository { return &memoryDashboardRepo{store: m} }
func (m *Memory) Audit() AuditRepository         { return &memoryAuditRepo{store: m} }
func (m *Memory) Relay() RelayRepository         { return m.relayRepo }
func (m *Memory) Credential() CredentialRepository {
	return &memoryCredentialRepo{store: m}
}
func (m *Memory) SchedulerLock() SchedulerLockRepository {
	return &memorySchedulerLockRepo{store: m}
}
func (m *Memory) Usage() UsageRepository {
	return &memoryUsageRepo{store: m}
}
func (m *Memory) Notification() NotificationRepository {
	return &memoryNotificationRepo{store: m}
}

func (m *Memory) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return deepCopySnapshot(m.data)
}

func (m *Memory) LoadSnapshot(snapshot Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = deepCopySnapshot(snapshot)
}

func deepCopySnapshot(snapshot Snapshot) Snapshot {
	return Snapshot{
		DataSourceStatus:    snapshot.DataSourceStatus,
		SyncConfig:          snapshot.SyncConfig,
		SyncLogs:            cloneSyncLogs(snapshot.SyncLogs),
		ImportFailures:      cloneImportFailures(snapshot.ImportFailures),
		Departments:         cloneDepartments(snapshot.Departments),
		Members:             cloneMembers(snapshot.Members),
		Roles:               cloneRoles(snapshot.Roles),
		Permissions:         clonePermissions(snapshot.Permissions),
		BudgetTree:          cloneBudgetTree(snapshot.BudgetTree),
		BudgetGroups:        cloneBudgetGroups(snapshot.BudgetGroups),
		OverrunPolicy:       snapshot.OverrunPolicy,
		AlertRules:          cloneAlertRules(snapshot.AlertRules),
		MemberQuotaPools:    cloneMemberQuotaPools(snapshot.MemberQuotaPools),
		ProviderKeys:        cloneProviderKeys(snapshot.ProviderKeys),
		PlatformKeys:        clonePlatformKeys(snapshot.PlatformKeys),
		Approvals:           cloneApprovals(snapshot.Approvals),
		Models:              cloneModels(snapshot.Models),
		RoutingRules:        cloneRoutingRules(snapshot.RoutingRules),
		ModelUsage:          cloneModelUsage(snapshot.ModelUsage),
		TeamUsage:           cloneTeamUsage(snapshot.TeamUsage),
		AuditSettings:       snapshot.AuditSettings,
		OperationLogs:       cloneOperationLogs(snapshot.OperationLogs),
		CallLogs:            cloneCallLogs(snapshot.CallLogs),
		CredentialPlatform:  snapshot.CredentialPlatform,
		EncryptedCredential: append([]byte(nil), snapshot.EncryptedCredential...),
	}
}
