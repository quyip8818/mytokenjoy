package memory

import (
	"sync"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
)

const defaultCompanyID int64 = 1

type Store struct {
	mu                   sync.RWMutex
	companies            map[int64]store.Company
	dataByCompany        map[int64]store.Snapshot
	permissions          []types.Permission
	providerKeys         []types.ProviderKey
	relayRepo            *memoryRelayRepo
	schedulerLocks       map[string]schedulerLockEntry
	usageBuckets         map[string]types.UsageBucketRow
	notificationLogs     []types.NotificationLogEntry
	invites              map[string]store.CompanyInvite
	platformOperators    map[string]store.PlatformOperator
	rechargeOrders       map[string]store.RechargeOrder
	memberPasswordHashes map[string]string
	logStore             store.LogStore
}

func New(snapshot store.Snapshot) *Store {
	m := &Store{
		companies:            make(map[int64]store.Company),
		dataByCompany:        make(map[int64]store.Snapshot),
		memberPasswordHashes: make(map[string]string),
	}
	m.initFromSnapshot(snapshot)
	m.applyDemoPasswords()
	m.logStore = newMemoryLogStore()
	m.relayRepo = newMemoryRelayRepo(m)
	return m
}

func (m *Store) initFromSnapshot(snapshot store.Snapshot) {
	company := snapshot.Company
	tid := company.ID
	m.companies[tid] = company
	snap := store.CloneSnapshot(snapshot)
	m.permissions = store.ClonePermissions(snap.Permissions)
	m.providerKeys = store.CloneProviderKeys(snap.ProviderKeys)
	m.dataByCompany[tid] = snap
}

func (m *Store) applyDemoPasswords() {
	hash := seed.DemoPasswordHash()
	for companyID, snap := range m.dataByCompany {
		for _, member := range snap.Members {
			if member.Status == "active" && member.Email != "" {
				m.memberPasswordHashes[memberPasswordKey(companyID, member.ID)] = hash
			}
		}
	}
}

func (m *Store) companySnapshot(companyID int64) store.Snapshot {
	snap, ok := m.dataByCompany[companyID]
	if !ok {
		return store.Snapshot{}
	}
	return snap
}

func (m *Store) setCompanySnapshot(companyID int64, snap store.Snapshot) {
	m.dataByCompany[companyID] = snap
}

func (m *Store) Company() store.CompanyRepository { return &memoryCompanyRepo{store: m} }
func (m *Store) Invite() store.InviteRepository   { return &memoryInviteRepo{store: m} }
func (m *Store) Platform() store.PlatformRepository {
	return &memoryPlatformRepo{store: m}
}
func (m *Store) Billing() store.BillingRepository { return &memoryBillingRepo{store: m} }
func (m *Store) Org() store.OrgRepository {
	return &memoryOrgRepo{store: m, nodes: &memoryOrgNodeRepo{store: m}}
}
func (m *Store) Budget() store.BudgetRepository { return &memoryBudgetRepo{store: m} }
func (m *Store) Keys() store.KeysRepository     { return &memoryKeysRepo{store: m} }
func (m *Store) Models() store.ModelsRepository {
	return &memoryModelsRepo{store: m, allowlist: &memoryModelAllowlistRepo{store: m}}
}
func (m *Store) Audit() store.AuditRepository   { return &memoryAuditRepo{store: m} }
func (m *Store) Ledger() store.LedgerRepository { return &memoryLedgerRepo{store: m} }
func (m *Store) Relay() store.RelayRepository   { return m.relayRepo }
func (m *Store) SchedulerLock() store.SchedulerLockRepository {
	return &memorySchedulerLockRepo{store: m}
}
func (m *Store) Usage() store.UsageRepository {
	return &memoryUsageRepo{store: m}
}
func (m *Store) Notification() store.NotificationRepository {
	return &memoryNotificationRepo{store: m}
}

func (m *Store) Snapshot() store.Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap := store.CloneSnapshot(m.companySnapshot(defaultCompanyID))
	if c, ok := m.companies[defaultCompanyID]; ok {
		snap.Company = c
	}
	snap.Permissions = store.ClonePermissions(m.permissions)
	snap.ProviderKeys = store.CloneProviderKeys(m.providerKeys)
	return snap
}

func (m *Store) LoadSnapshot(snapshot store.Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.companies = make(map[int64]store.Company)
	m.dataByCompany = make(map[int64]store.Snapshot)
	m.memberPasswordHashes = make(map[string]string)
	m.permissions = nil
	m.providerKeys = nil
	m.initFromSnapshot(snapshot)
	m.applyDemoPasswords()
}

func (m *Store) RelayOutboxEntry(id string) (store.RelayOutboxEntry, bool) {
	return m.relayRepo.relayOutboxEntry(id)
}

func (m *Store) MemberPasswordHash(companyID int64, memberID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.memberPasswordHashes == nil {
		return "", false
	}
	hash, ok := m.memberPasswordHashes[memberPasswordKey(companyID, memberID)]
	return hash, ok
}

var _ store.Store = (*Store)(nil)
