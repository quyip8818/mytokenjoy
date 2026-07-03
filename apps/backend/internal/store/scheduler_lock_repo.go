package store

import (
	"context"
	"time"
)

type SchedulerLockRepository interface {
	TryAcquire(ctx context.Context, lockName, holder string, lease time.Duration) (bool, error)
	Release(ctx context.Context, lockName, holder string) error
}
