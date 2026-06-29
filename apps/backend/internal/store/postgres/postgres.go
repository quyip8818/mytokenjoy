package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
)

type Store struct {
	pool   *pgxpool.Pool
	relay  *relayRepo
	domain domainRepos
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
	if err := applySchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	s := &Store{pool: pool}
	if err := s.loadOrSeedDomain(ctx, cfg); err != nil {
		pool.Close()
		return nil, err
	}
	if cfg.IsDemoProfile() {
		if err := seed.ApplyUsageBuckets(ctx, s, cfg); err != nil {
			pool.Close()
			return nil, err
		}
	}
	s.relay = newRelayRepo(ctx, pool)
	s.domain = newDomainRepoSet(newPoolShardBackend(ctx, pool))
	return s, nil
}

func newDomainRepoSet(backend shardBackend) domainRepos {
	org, budget, keys, models, audit := newDomainRepos(backend)
	return domainRepos{
		org: org, budget: budget, keys: keys, models: models, audit: audit,
	}
}

func (s *Store) Credential() store.CredentialRepository {
	return &credentialRepo{db: s.pool}
}

func (s *Store) SchedulerLock() store.SchedulerLockRepository {
	return &schedulerLockRepo{db: s.pool}
}

func (s *Store) Usage() store.UsageRepository {
	return &usageRepo{db: s.pool}
}

func (s *Store) Notification() store.NotificationRepository {
	return &notificationRepo{db: s.pool}
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Org() store.OrgRepository       { return s.domain.org }
func (s *Store) Budget() store.BudgetRepository { return s.domain.budget }
func (s *Store) Keys() store.KeysRepository     { return s.domain.keys }
func (s *Store) Models() store.ModelsRepository { return s.domain.models }
func (s *Store) Audit() store.AuditRepository   { return s.domain.audit }
func (s *Store) Relay() store.RelayRepository   { return s.relay }

func (s *Store) loadOrSeedDomain(ctx context.Context, cfg config.Config) error {
	shards, err := loadAllShards(ctx, s.pool)
	if err != nil {
		return err
	}
	if shardsComplete(shards) {
		return nil
	}
	if cfg.IsProdProfile() {
		return fmt.Errorf("postgres domain shards incomplete in prod profile")
	}
	return seedShards(ctx, s.pool, seed.Load(cfg))
}
