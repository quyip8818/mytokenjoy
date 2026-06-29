package postgres

import (
	"context"
	"embed"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	pool        *pgxpool.Pool
	memory      *store.Memory
	relay       *relayRepo
	activeCtx   context.Context
	activeMu    sync.RWMutex
	dirtyShards map[string]struct{}
	dirtyMu     sync.Mutex
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
	if err := migrate(ctx, pool); err != nil {
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
	s.relay = &relayRepo{db: pool}
	return s, nil
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

func (s *Store) Org() store.OrgRepository {
	return &persistOrgRepo{inner: s.memory.Org(), store: s}
}
func (s *Store) Budget() store.BudgetRepository {
	return &persistBudgetRepo{inner: s.memory.Budget(), store: s}
}
func (s *Store) Keys() store.KeysRepository {
	return &persistKeysRepo{inner: s.memory.Keys(), store: s}
}
func (s *Store) Models() store.ModelsRepository {
	return &persistModelsRepo{inner: s.memory.Models(), store: s}
}
func (s *Store) Audit() store.AuditRepository {
	return &persistAuditRepo{inner: s.memory.Audit(), store: s}
}
func (s *Store) Relay() store.RelayRepository { return s.relay }

func (s *Store) loadOrSeedDomain(ctx context.Context, cfg config.Config) error {
	shards, err := s.loadShards(ctx)
	if err != nil {
		return err
	}
	if shardsComplete(shards) {
		snapshot, err := store.ShardsToSnapshot(shards)
		if err != nil {
			return err
		}
		s.memory = store.NewMemory(snapshot)
		return nil
	}
	if cfg.IsProdProfile() {
		return fmt.Errorf("postgres domain shards incomplete in prod profile")
	}
	return s.seedDomainShards(ctx, seed.Load(cfg))
}

func (s *Store) domainPersistCtx() context.Context {
	s.activeMu.RLock()
	defer s.activeMu.RUnlock()
	if s.activeCtx != nil {
		return s.activeCtx
	}
	return context.Background()
}

func (s *Store) setActiveCtx(ctx context.Context) {
	s.activeMu.Lock()
	s.activeCtx = ctx
	s.activeMu.Unlock()
}

func (s *Store) clearActiveCtx() {
	s.activeMu.Lock()
	s.activeCtx = nil
	s.activeMu.Unlock()
}
