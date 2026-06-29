package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type txStore struct {
	domain       domainRepos
	relay        store.RelayRepository
	usage        store.UsageRepository
	notification store.NotificationRepository
	parent       *Store
}

func (s *txStore) Org() store.OrgRepository       { return s.domain.org }
func (s *txStore) Budget() store.BudgetRepository { return s.domain.budget }
func (s *txStore) Keys() store.KeysRepository     { return s.domain.keys }
func (s *txStore) Models() store.ModelsRepository { return s.domain.models }
func (s *txStore) Audit() store.AuditRepository   { return s.domain.audit }

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

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txView := &txStore{
		domain:       newDomainRepoSet(ctx, tx),
		relay:        newRelayRepo(ctx, tx),
		usage:        &usageRepo{db: tx},
		notification: &notificationRepo{db: tx},
		parent:       s,
	}
	if err := fn(txView); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

var _ store.Store = (*Store)(nil)
var _ store.Store = (*txStore)(nil)
