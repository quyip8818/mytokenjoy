package store

import (
	"fmt"
	"sync"
	"time"
)

type memoryRelayRepo struct {
	store *Memory
	mu    sync.Mutex
	data  struct {
		mappings      map[string]RelayMapping
		tokenIndex    map[int64]string
		relayOutbox   []RelayOutboxEntry
		webhookOutbox []WebhookOutboxEntry
		ingestedLogs  map[int64]struct{}
		lastLogID     int64
		rebalance     []RebalanceQueueEntry
	}
}

func newMemoryRelayRepo(m *Memory) *memoryRelayRepo {
	r := &memoryRelayRepo{store: m}
	r.data.mappings = make(map[string]RelayMapping)
	r.data.tokenIndex = make(map[int64]string)
	r.data.ingestedLogs = make(map[int64]struct{})
	return r
}

func (r *memoryRelayRepo) GetMappingByPlatformKeyID(platformKeyID string) (*RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok {
		return nil, nil
	}
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) GetMappingByNewAPITokenID(tokenID int64) (*RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.data.tokenIndex[tokenID]
	if !ok {
		return nil, nil
	}
	m := r.data.mappings[id]
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) ListMappingsByMemberID(memberID string) ([]RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RelayMapping, 0)
	for _, m := range r.data.mappings {
		if m.MemberID != nil && *m.MemberID == memberID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByDepartmentID(departmentID string) ([]RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RelayMapping, 0)
	for _, m := range r.data.mappings {
		if m.DepartmentID == departmentID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByBudgetGroupID(budgetGroupID string) ([]RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RelayMapping, 0)
	for _, m := range r.data.mappings {
		if m.BudgetGroupID != nil && *m.BudgetGroupID == budgetGroupID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListActiveMappings() ([]RelayMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RelayMapping, 0)
	for _, m := range r.data.mappings {
		if m.SyncStatus == RelaySyncStatusSynced {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) UpsertMapping(mapping RelayMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.mappings[mapping.PlatformKeyID] = mapping
	if mapping.NewAPITokenID != nil {
		r.data.tokenIndex[*mapping.NewAPITokenID] = mapping.PlatformKeyID
	}
	return nil
}

func (r *memoryRelayRepo) UpdateMappingSync(
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.NewAPITokenID = &tokenID
	m.SyncStatus = status
	m.SyncedAt = &syncedAt
	m.RelayRemainQuota = remainQuota
	r.data.mappings[platformKeyID] = m
	r.data.tokenIndex[tokenID] = platformKeyID
	return nil
}

func (r *memoryRelayRepo) UpdateMappingRemainQuota(platformKeyID string, remainQuota int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.RelayRemainQuota = &remainQuota
	r.data.mappings[platformKeyID] = m
	return nil
}

func (r *memoryRelayRepo) EnqueueRelayOutbox(entry RelayOutboxEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.relayOutbox = append(r.data.relayOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRelayOutbox(limit int) ([]RelayOutboxEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]RelayOutboxEntry, 0, limit)
	for i := range r.data.relayOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.relayOutbox[i]
		if e.Status == OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRelayOutboxDone(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Status = OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkRelayOutboxRetry(id string, nextRetry time.Time, lastError string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Attempts++
			r.data.relayOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.relayOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) EnqueueWebhookOutbox(entry WebhookOutboxEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.webhookOutbox = append(r.data.webhookOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingWebhookOutbox(limit int) ([]WebhookOutboxEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]WebhookOutboxEntry, 0, limit)
	for i := range r.data.webhookOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.webhookOutbox[i]
		if e.Status == OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxDone(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Status = OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxRetry(id string, nextRetry time.Time, lastError string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Attempts++
			r.data.webhookOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.webhookOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) HasIngestedLogID(logID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data.ingestedLogs[logID]
	return ok, nil
}

func (r *memoryRelayRepo) InsertIngestedLogID(logID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.ingestedLogs[logID] = struct{}{}
	return nil
}

func (r *memoryRelayRepo) GetLastLogID() (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data.lastLogID, nil
}

func (r *memoryRelayRepo) SetLastLogID(logID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.lastLogID = logID
	return nil
}

func (r *memoryRelayRepo) EnqueueRebalance(axisKind, axisID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range r.data.rebalance {
		if e.AxisKind == axisKind && e.AxisID == axisID && e.Status == OutboxStatusPending {
			return nil
		}
	}
	r.data.rebalance = append(r.data.rebalance, RebalanceQueueEntry{
		ID:       fmt.Sprintf("rb-%s-%s-%d", axisKind, axisID, time.Now().UnixNano()),
		AxisKind: axisKind,
		AxisID:   axisID,
		Status:   OutboxStatusPending,
	})
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRebalance(limit int) ([]RebalanceQueueEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RebalanceQueueEntry, 0, limit)
	for _, e := range r.data.rebalance {
		if len(out) >= limit {
			break
		}
		if e.Status == OutboxStatusPending {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRebalanceDone(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.rebalance {
		if r.data.rebalance[i].ID == id {
			r.data.rebalance[i].Status = OutboxStatusDone
			return nil
		}
	}
	return nil
}
