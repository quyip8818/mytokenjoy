package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSeedBudgetConsumedAlignWithClockAnchor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaURL := testutil.TestSchemaURL(t)
	cfg := testutil.PreparedConfig(schemaURL)
	cfg.ClockAnchor = "2026-06-19"

	st, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}
	if pg, ok := st.(*postgres.Store); ok {
		pg.Close()
	}

	pool, err := pgxpool.New(ctx, schemaURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := truncateDomainTables(ctx, pool); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	snap := seed.Load(cfg)
	if snap.SeedAt.IsZero() {
		t.Fatal("expected SeedAt from clock")
	}
	wantPeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if wantPeriod != "2026-06" {
		t.Fatalf("want open period 2026-06, got %q", wantPeriod)
	}
	if !snap.SeedAt.Equal(cfg.Clock().Now().UTC()) {
		t.Fatalf("SeedAt=%v want clock %v", snap.SeedAt, cfg.Clock().Now().UTC())
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.ApplyTables(ctx, tx, snap); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	rows, err := pool.Query(ctx, `
		SELECT DISTINCT period_key FROM budget_consumed WHERE company_id = $1
	`, contract.LocalCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("expected seeded budget_consumed")
	}
	for _, key := range keys {
		if key != wantPeriod {
			t.Fatalf("budget_consumed period_key=%q want %q", key, wantPeriod)
		}
	}

	ledgerRows, err := pool.Query(ctx, `
		SELECT period_key, occurred_at FROM usage_ledger WHERE company_id = $1
	`, contract.LocalCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	defer ledgerRows.Close()
	for ledgerRows.Next() {
		var periodKey string
		var occurredAt time.Time
		if err := ledgerRows.Scan(&periodKey, &occurredAt); err != nil {
			t.Fatal(err)
		}
		wantLedger := pkgbudget.OccurrenceSnapshotKey(pkgbudget.PeriodMonthly, occurredAt.UTC()).String()
		if periodKey != wantLedger {
			t.Fatalf("usage_ledger period_key=%q want %q from occurred_at=%v", periodKey, wantLedger, occurredAt)
		}
	}
	if err := ledgerRows.Err(); err != nil {
		t.Fatal(err)
	}
}
