package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryAuditRepo struct{ store *Store }

func (r *memoryAuditRepo) Settings() types.AuditSettings {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.AuditSettings
}

func (r *memoryAuditRepo) SetSettings(settings types.AuditSettings) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AuditSettings = settings
	return nil
}

func (r *memoryAuditRepo) OperationLogs() []types.OperationLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneOperationLogs(r.store.data.OperationLogs)
}

func (r *memoryAuditRepo) CallLogs() []types.CallLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneCallLogs(r.store.data.CallLogs)
}
