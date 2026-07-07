package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryLogStore struct {
	mu       sync.Mutex
	logs     map[int64]store.RawConsumeLog
	cursor   int64
	failures map[string]store.IngestFailure
	byLogID  map[int64]string
}

func newMemoryLogStore() *memoryLogStore {
	return &memoryLogStore{
		logs:     make(map[int64]store.RawConsumeLog),
		failures: make(map[string]store.IngestFailure),
		byLogID:  make(map[int64]string),
	}
}

func (m *memoryLogStore) PutConsumeLog(raw store.RawConsumeLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs[raw.ID] = raw
}

func (m *memoryLogStore) GetConsumeLogByID(_ context.Context, logID int64) (*store.RawConsumeLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, ok := m.logs[logID]
	if !ok {
		return nil, store.ErrConsumeLogNotFound
	}
	copy := raw
	return &copy, nil
}

func (m *memoryLogStore) ListConsumeLogIDsAfter(_ context.Context, afterID int64, limit int) ([]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]int64, 0, len(m.logs))
	for id := range m.logs {
		if id > afterID && m.logs[id].TokenID > 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}
	return ids, nil
}

func (m *memoryLogStore) GetReconcileCursor(_ context.Context, _ string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cursor, nil
}

func (m *memoryLogStore) SetReconcileCursor(_ context.Context, _ string, logID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursor = logID
	return nil
}

func (m *memoryLogStore) UpsertFailure(_ context.Context, f store.IngestFailure) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if f.ID == "" {
		f.ID = store.IngestFailureID(f.LogID)
	}
	now := time.Now().UTC()
	if existingID, ok := m.byLogID[f.LogID]; ok {
		existing := m.failures[existingID]
		existing.Source = f.Source
		existing.Error = f.Error
		existing.UpdatedAt = now
		m.failures[existingID] = existing
		return nil
	}
	if f.Status == "" {
		f.Status = store.IngestFailureStatusPending
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	if f.NextRetry.IsZero() {
		f.NextRetry = now
	}
	f.UpdatedAt = now
	m.failures[f.ID] = f
	m.byLogID[f.LogID] = f.ID
	return nil
}

func (m *memoryLogStore) ClaimPendingFailures(_ context.Context, limit int) ([]store.IngestFailure, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	leaseUntil := now.Add(store.FailureClaimLease())
	candidates := make([]string, 0)
	for id, f := range m.failures {
		if f.Status != store.IngestFailureStatusPending || f.Attempts >= store.IngestFailureMaxAttempts {
			continue
		}
		if f.NextRetry.After(now) {
			continue
		}
		candidates = append(candidates, id)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return m.failures[candidates[i]].NextRetry.Before(m.failures[candidates[j]].NextRetry)
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]store.IngestFailure, 0, len(candidates))
	for _, id := range candidates {
		f := m.failures[id]
		f.NextRetry = leaseUntil
		f.UpdatedAt = now
		m.failures[id] = f
		out = append(out, f)
	}
	return out, nil
}

func (m *memoryLogStore) MarkFailureDone(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.failures[id]
	if ok {
		delete(m.byLogID, f.LogID)
		delete(m.failures, id)
	}
	return nil
}

func (m *memoryLogStore) MarkFailureRetry(_ context.Context, id string, next time.Time, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.failures[id]
	if !ok {
		return nil
	}
	f.Attempts++
	f.NextRetry = next
	f.Error = errMsg
	f.UpdatedAt = time.Now().UTC()
	m.failures[id] = f
	return nil
}

func (m *memoryLogStore) MarkFailureDead(_ context.Context, id string, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.failures[id]
	if !ok {
		return nil
	}
	f.Status = store.IngestFailureStatusDead
	f.Error = errMsg
	f.UpdatedAt = time.Now().UTC()
	m.failures[id] = f
	return nil
}

func (m *memoryLogStore) CountConsumeLogsAfter(_ context.Context, afterID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for id, raw := range m.logs {
		if id > afterID && raw.TokenID > 0 {
			count++
		}
	}
	return count, nil
}

func (m *memoryLogStore) CountPendingIngestFailures(_ context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, f := range m.failures {
		if f.Status == store.IngestFailureStatusPending && f.Attempts < store.IngestFailureMaxAttempts {
			count++
		}
	}
	return count, nil
}

func (m *memoryLogStore) IngestLagSeconds(_ context.Context, afterID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var oldest *int64
	for id, raw := range m.logs {
		if id <= afterID || raw.TokenID <= 0 {
			continue
		}
		if oldest == nil || raw.CreatedAt < *oldest {
			v := raw.CreatedAt
			oldest = &v
		}
	}
	if oldest == nil {
		return 0, nil
	}
	lag := time.Now().Unix() - *oldest
	if lag < 0 {
		return 0, nil
	}
	return lag, nil
}

func (m *memoryLogStore) setFailureNextRetry(id string, next time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.failures[id]
	if !ok {
		return
	}
	f.NextRetry = next
	m.failures[id] = f
}

func (m *memoryLogStore) ingestFailureByLogID(logID int64) (store.IngestFailure, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byLogID[logID]
	if !ok {
		return store.IngestFailure{}, false
	}
	f, ok := m.failures[id]
	if !ok {
		return store.IngestFailure{}, false
	}
	return f, true
}

func (m *Store) SetIngestFailureNextRetry(id string, next time.Time) {
	if mem, ok := m.logStore.(*memoryLogStore); ok {
		mem.setFailureNextRetry(id, next)
	}
}

func (m *Store) IngestFailureByLogID(logID int64) (store.IngestFailure, bool) {
	mem, ok := m.logStore.(*memoryLogStore)
	if !ok {
		return store.IngestFailure{}, false
	}
	return mem.ingestFailureByLogID(logID)
}

func (m *Store) Logs() store.LogStore {
	return m.logStore
}

func (m *Store) PutConsumeLog(raw store.RawConsumeLog) {
	if mem, ok := m.logStore.(*memoryLogStore); ok {
		mem.PutConsumeLog(raw)
	}
}
