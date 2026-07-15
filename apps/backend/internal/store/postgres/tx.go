package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type txStore struct {
	tx                          pgx.Tx
	domain                      domainRepos
	ledger                      store.LedgerRepository
	mappings                    store.PlatformKeyMappingRepository
	gatewayPrecheck             store.GatewayPrecheckRepository
	combinedKeySummaries        store.CombinedKeySummaryRepository
	budgetConsumed              store.BudgetConsumedRepository
	dashboardProjectionProgress store.ProjectionProgressRepository
	usage                       store.UsageRepository
	notification                store.NotificationRepository
	company                     store.CompanyRepository
	invite                      store.InviteRepository
	billing                     store.BillingRepository
	tenantBackgroundState       store.TenantBackgroundStateRepository
	riverJob                    store.RiverJobRepository
	parent                      *Store
}

func (s *txStore) PgxTx() pgx.Tx { return s.tx }

func (s *txStore) Org() store.OrgRepository       { return s.domain.org }
func (s *txStore) Budget() store.BudgetRepository { return s.domain.budget }
func (s *txStore) Keys() store.KeysRepository     { return s.domain.keys }
func (s *txStore) Models() store.ModelsRepository { return s.domain.models }
func (s *txStore) Audit() store.AuditRepository   { return s.domain.audit }
func (s *txStore) Ledger() store.LedgerRepository { return s.ledger }

func (s *txStore) BudgetConsumed() store.BudgetConsumedRepository {
	return s.budgetConsumed
}

func (s *txStore) DashboardProjectionProgress() store.ProjectionProgressRepository {
	return s.dashboardProjectionProgress
}

func (s *txStore) PlatformKeyMappings() store.PlatformKeyMappingRepository {
	return s.mappings
}

func (s *txStore) GatewayPrecheck() store.GatewayPrecheckRepository {
	return s.gatewayPrecheck
}

func (s *txStore) CombinedKeySummaries() store.CombinedKeySummaryRepository {
	return s.combinedKeySummaries
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

func (s *txStore) NotificationPreference() store.NotificationPreferenceRepository {
	return &notificationPreferenceRepo{db: s.tx}
}

func (s *txStore) Company() store.CompanyRepository {
	return s.company
}

func (s *txStore) Invite() store.InviteRepository {
	return s.invite
}

func (s *txStore) Platform() store.PlatformRepository {
	return s.parent.Platform()
}

func (s *txStore) Billing() store.BillingRepository {
	return s.billing
}

func (s *txStore) TenantBackgroundState() store.TenantBackgroundStateRepository {
	return s.tenantBackgroundState
}

func (s *txStore) RiverJob() store.RiverJobRepository {
	return s.riverJob
}

func (s *txStore) Logs() store.LogStore {
	return s.parent.Logs()
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
		tx:                          tx,
		domain:                      newDomainRepoSet(tx, s.tokenJoyCompanyID, s.credentialKey),
		ledger:                      &pgLedgerRepo{db: tx},
		mappings:                    newPlatformKeyMappingRepo(tx),
		gatewayPrecheck:             newGatewayPrecheckRepo(tx),
		combinedKeySummaries:        newCombinedKeySummaryRepo(tx),
		budgetConsumed:              newBudgetConsumedRepo(tx),
		dashboardProjectionProgress: newDashboardProjectionProgressRepo(tx),
		usage:                       &usageRepo{db: tx},
		notification:                &notificationRepo{db: tx},
		company:                     newCompanyRepo(tx),
		invite:                      newInviteRepo(tx),
		billing:                     newBillingRepo(tx),
		tenantBackgroundState:       newTenantBackgroundStateRepo(tx),
		riverJob:                    newRiverJobRepo(tx),
		parent:                      s,
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
var _ store.Tx = (*txStore)(nil)
