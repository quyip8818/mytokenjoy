package memory

import (
	"sync"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Store struct {
	mu               sync.RWMutex
	data             store.Snapshot
	relayRepo        *memoryRelayRepo
	schedulerLocks   map[string]schedulerLockEntry
	usageBuckets     map[string]types.UsageBucketRow
	notificationLogs []types.NotificationLogEntry
}

func New(snapshot store.Snapshot) *Store {
	m := &Store{data: store.CloneSnapshot(snapshot)}
	m.relayRepo = newMemoryRelayRepo(m)
	return m
}

func (m *Store) Org() store.OrgRepository       { return &memoryOrgRepo{store: m} }
func (m *Store) Budget() store.BudgetRepository { return &memoryBudgetRepo{store: m} }
func (m *Store) Keys() store.KeysRepository     { return &memoryKeysRepo{store: m} }
func (m *Store) Models() store.ModelsRepository { return &memoryModelsRepo{store: m} }
func (m *Store) Audit() store.AuditRepository   { return &memoryAuditRepo{store: m} }
func (m *Store) Relay() store.RelayRepository   { return m.relayRepo }
func (m *Store) Credential() store.CredentialRepository {
	return &memoryCredentialRepo{store: m}
}
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
	return store.CloneSnapshot(m.data)
}

func (m *Store) LoadSnapshot(snapshot store.Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = store.CloneSnapshot(snapshot)
}

func (m *Store) RelayOutboxEntry(id string) (store.RelayOutboxEntry, bool) {
	return m.relayRepo.relayOutboxEntry(id)
}

var _ store.Store = (*Store)(nil)
