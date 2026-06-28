package postgres

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	pool   *pgxpool.Pool
	memory *store.Memory
	relay  *relayRepo
}

func New(ctx context.Context, databaseURL string, seed store.Snapshot) (store.Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
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
	if err := s.loadOrSeedDomain(ctx, seed); err != nil {
		pool.Close()
		return nil, err
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
	return &persistOrgRepo{inner: s.memory.Org(), persist: s.persistDomain}
}
func (s *Store) Budget() store.BudgetRepository {
	return &persistBudgetRepo{inner: s.memory.Budget(), persist: s.persistDomain}
}
func (s *Store) Keys() store.KeysRepository {
	return &persistKeysRepo{inner: s.memory.Keys(), persist: s.persistDomain}
}
func (s *Store) Models() store.ModelsRepository {
	return &persistModelsRepo{inner: s.memory.Models(), persist: s.persistDomain}
}
func (s *Store) Dashboard() store.DashboardRepository { return s.memory.Dashboard() }
func (s *Store) Audit() store.AuditRepository {
	return &persistAuditRepo{inner: s.memory.Audit(), persist: s.persistDomain}
}
func (s *Store) Relay() store.RelayRepository { return s.relay }

func (s *Store) loadOrSeedDomain(ctx context.Context, seed store.Snapshot) error {
	var raw []byte
	err := s.pool.QueryRow(ctx, `SELECT data FROM domain_snapshot WHERE id = 1`).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		s.memory = store.NewMemory(seed)
		return s.persistDomain(ctx)
	}
	if err != nil {
		return fmt.Errorf("load domain snapshot: %w", err)
	}
	var snapshot store.Snapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return fmt.Errorf("unmarshal domain snapshot: %w", err)
	}
	s.memory = store.NewMemory(snapshot)
	return nil
}

func (s *Store) persistDomain(ctx context.Context) error {
	raw, err := json.Marshal(s.memory.Snapshot())
	if err != nil {
		return fmt.Errorf("marshal domain snapshot: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO domain_snapshot (id, data, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()
	`, raw)
	if err != nil {
		return fmt.Errorf("persist domain snapshot: %w", err)
	}
	return nil
}
