package postgres

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type schedulerLockRepo struct {
	db dbQuerier
}

func (r *schedulerLockRepo) TryAcquire(
	ctx context.Context,
	lockName, holder string,
	lease time.Duration,
) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO scheduler_locks (lock_name, holder, lease_until, updated_at)
		VALUES ($1, $2, NOW() + ($3::bigint * interval '1 microsecond'), NOW())
		ON CONFLICT (lock_name) DO UPDATE
		SET holder = EXCLUDED.holder,
		    lease_until = EXCLUDED.lease_until,
		    updated_at = NOW()
		WHERE scheduler_locks.lease_until < NOW() OR scheduler_locks.holder = EXCLUDED.holder
	`, lockName, holder, lease.Microseconds())
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *schedulerLockRepo) Release(ctx context.Context, lockName, holder string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM scheduler_locks
		WHERE lock_name = $1 AND holder = $2
	`, lockName, holder)
	return err
}

var _ store.SchedulerLockRepository = (*schedulerLockRepo)(nil)
