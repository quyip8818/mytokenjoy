package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/runtime"
)

type Store struct {
	pool                 *pgxpool.Pool
	logPool              *pgxpool.Pool
	logTables            logTables
	mappings             *platformKeyMappingRepo
	gatewayPrecheck      *gatewayPrecheckRepo
	gatewaySoftSummaries *gatewaySoftSummaryRepo
	logs                 store.LogStore
	domain               domainRepos
	tokenJoyCompanyID    int64
	credentialKey        []byte
}

type domainRepos struct {
	org    store.OrgRepository
	budget store.BudgetRepository
	keys   store.KeysRepository
	models store.ModelsRepository
	audit  store.AuditRepository
}

func New(ctx context.Context, cfg config.Config) (store.Store, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if !cfg.StoreBootstrap.SchemaPrepared {
		if err := applySchema(ctx, pool, cfg); err != nil {
			pool.Close()
			return nil, err
		}
		if err := ensureBootstrapCompany(ctx, pool, cfg); err != nil {
			pool.Close()
			return nil, err
		}
	}

	credentialKey, err := common.ParseKey(cfg.DataSourceCredentialKey)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("parse DATA_SOURCE_CREDENTIAL_KEY: %w", err)
	}
	s := &Store{
		pool:              pool,
		logs:              store.NoopLogStore(),
		tokenJoyCompanyID: cfg.TokenJoyCompanyID,
		credentialKey:     credentialKey,
	}
	if cfg.IngestEnabled() {
		tables := logTablesFor(cfg)
		logPool, err := pgxpool.New(ctx, cfg.LogDatabaseURL)
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("connect logs postgres: %w", err)
		}
		if err := logPool.Ping(ctx); err != nil {
			logPool.Close()
			pool.Close()
			return nil, fmt.Errorf("ping logs postgres: %w", err)
		}
		if err := applyLogsSchema(ctx, logPool, cfg); err != nil {
			logPool.Close()
			pool.Close()
			return nil, err
		}
		s.logPool = logPool
		s.logTables = tables
		s.logs = newLogRepo(logPool, tables)
	}
	if !cfg.StoreBootstrap.SchemaPrepared {
		if err := s.loadOrSeedDomain(ctx, cfg); err != nil {
			pool.Close()
			if s.logPool != nil {
				s.logPool.Close()
			}
			return nil, err
		}
	}
	s.mappings = newPlatformKeyMappingRepo(pool)
	s.gatewayPrecheck = newGatewayPrecheckRepo(pool)
	s.gatewaySoftSummaries = newGatewaySoftSummaryRepo(pool)
	s.domain = newDomainRepoSet(pool, s.tokenJoyCompanyID, s.credentialKey)
	return s, nil
}

func (s *Store) Company() store.CompanyRepository {
	return newCompanyRepo(s.pool)
}

func (s *Store) Invite() store.InviteRepository {
	return newInviteRepo(s.pool)
}

func (s *Store) Platform() store.PlatformRepository {
	return newPlatformRepo(s.pool)
}

func (s *Store) Billing() store.BillingRepository {
	return newBillingRepo(s.pool)
}

func (s *Store) SchedulerLock() store.SchedulerLockRepository {
	return &schedulerLockRepo{db: s.pool}
}

func (s *Store) TenantBackgroundState() store.TenantBackgroundStateRepository {
	return newTenantBackgroundStateRepo(s.pool)
}

func (s *Store) RiverJob() store.RiverJobRepository {
	return newRiverJobRepo(s.pool)
}

func (s *Store) Usage() store.UsageRepository {
	return &usageRepo{db: s.pool}
}

func (s *Store) Notification() store.NotificationRepository {
	return &notificationRepo{db: s.pool}
}

func (s *Store) Close() {
	if s.logPool != nil {
		s.logPool.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Org() store.OrgRepository       { return s.domain.org }
func (s *Store) Budget() store.BudgetRepository { return s.domain.budget }
func (s *Store) Keys() store.KeysRepository     { return s.domain.keys }
func (s *Store) Models() store.ModelsRepository { return s.domain.models }
func (s *Store) Audit() store.AuditRepository   { return s.domain.audit }
func (s *Store) Ledger() store.LedgerRepository { return &pgLedgerRepo{db: s.pool} }
func (s *Store) BudgetConsumed() store.BudgetConsumedRepository {
	return newBudgetConsumedRepo(s.pool)
}

func (s *Store) BudgetProjectionProgress() store.ProjectionProgressRepository {
	return newBudgetProjectionProgressRepo(s.pool)
}

func (s *Store) DashboardProjectionProgress() store.ProjectionProgressRepository {
	return newDashboardProjectionProgressRepo(s.pool)
}

func (s *Store) PlatformKeyMappings() store.PlatformKeyMappingRepository { return s.mappings }
func (s *Store) GatewayPrecheck() store.GatewayPrecheckRepository        { return s.gatewayPrecheck }
func (s *Store) GatewaySoftSummaries() store.GatewaySoftSummaryRepository {
	return s.gatewaySoftSummaries
}
func (s *Store) Logs() store.LogStore { return s.logs }

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) loadOrSeedDomain(ctx context.Context, cfg config.Config) error {
	empty, err := isDatabaseEmpty(ctx, s.pool)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	if !cfg.BootstrapIsMinimal() && !cfg.BootstrapIsDemo() {
		return fmt.Errorf("database empty: set BOOTSTRAP_MODE=minimal|demo or run seed")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)
	var snap store.Snapshot
	if cfg.BootstrapIsMinimal() {
		snap = seed.LoadMinimal(cfg)
	} else {
		snap = seed.Load(cfg)
	}
	if err := seed.ApplyTables(ctx, tx, snap); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	if cfg.BootstrapIsDemo() {
		return runtime.ApplyDemo(ctx, s, cfg)
	}
	return nil
}

func isDatabaseEmpty(ctx context.Context, exec dbQuerier) (bool, error) {
	var count int
	err := exec.QueryRow(ctx, `SELECT COUNT(*) FROM members`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count members: %w", err)
	}
	return count == 0, nil
}
