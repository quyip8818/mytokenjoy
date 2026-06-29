package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryNotificationRepo struct {
	store *Store
}

func (r *memoryNotificationRepo) Append(ctx context.Context, entry types.NotificationLogEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.notificationLogs = append(r.store.notificationLogs, entry)
	return nil
}

func (m *Store) NotificationLogs() []types.NotificationLogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]types.NotificationLogEntry, len(m.notificationLogs))
	copy(result, m.notificationLogs)
	return result
}

var _ store.NotificationRepository = (*memoryNotificationRepo)(nil)
