package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type txStore struct {
	memory       *store.Memory
	relay        store.RelayRepository
	usage        store.UsageRepository
	notification store.NotificationRepository
	parent       *Store
}

func (s *txStore) Org() store.OrgRepository {
	return &deferredOrgRepo{inner: s.memory.Org(), onMutate: s.parent.markDirty}
}

func (s *txStore) Budget() store.BudgetRepository {
	return &deferredBudgetRepo{inner: s.memory.Budget(), onMutate: s.parent.markDirty}
}

func (s *txStore) Keys() store.KeysRepository {
	return &deferredKeysRepo{inner: s.memory.Keys(), onMutate: s.parent.markDirty}
}

func (s *txStore) Models() store.ModelsRepository {
	return &deferredModelsRepo{inner: s.memory.Models(), onMutate: s.parent.markDirty}
}

func (s *txStore) Audit() store.AuditRepository {
	return &deferredAuditRepo{inner: s.memory.Audit(), onMutate: s.parent.markDirty}
}

func (s *txStore) Relay() store.RelayRepository {
	return s.relay
}

func (s *txStore) Credential() store.CredentialRepository {
	return s.parent.Credential()
}

func (s *txStore) SchedulerLock() store.SchedulerLockRepository {
	return s.parent.SchedulerLock()
}

func (s *txStore) Usage() store.UsageRepository {
	return s.usage
}

func (s *txStore) Notification() store.NotificationRepository {
	return s.notification
}

func (s *txStore) WithTx(ctx context.Context, fn func(store.Store) error) error {
	return fn(s)
}

func (s *Store) WithTx(ctx context.Context, fn func(store.Store) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.setActiveCtx(ctx)
	defer s.clearActiveCtx()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txView := &txStore{
		memory:       s.memory,
		relay:        &relayRepo{db: tx},
		usage:        &usageRepo{db: tx},
		notification: &notificationRepo{db: tx},
		parent:       s,
	}
	if err := fn(txView); err != nil {
		return err
	}
	if err := s.persistDomainExec(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Store) persistDomainExec(ctx context.Context, exec dbQuerier) error {
	return s.flushShards(ctx, exec)
}

var _ store.Store = (*Store)(nil)
var _ store.Store = (*txStore)(nil)
