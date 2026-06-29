package memory

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type schedulerLockEntry struct {
	holder     string
	leaseUntil time.Time
}

type memorySchedulerLockRepo struct {
	store *Store
}

func (r *memorySchedulerLockRepo) TryAcquire(
	ctx context.Context,
	lockName, holder string,
	lease time.Duration,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.schedulerLocks == nil {
		r.store.schedulerLocks = make(map[string]schedulerLockEntry)
	}
	now := time.Now()
	entry, ok := r.store.schedulerLocks[lockName]
	if ok && entry.leaseUntil.After(now) && entry.holder != holder {
		return false, nil
	}
	r.store.schedulerLocks[lockName] = schedulerLockEntry{
		holder:     holder,
		leaseUntil: now.Add(lease),
	}
	return true, nil
}

func (r *memorySchedulerLockRepo) Release(ctx context.Context, lockName, holder string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	entry, ok := r.store.schedulerLocks[lockName]
	if !ok || entry.holder != holder {
		return nil
	}
	delete(r.store.schedulerLocks, lockName)
	return nil
}

var _ store.SchedulerLockRepository = (*memorySchedulerLockRepo)(nil)
