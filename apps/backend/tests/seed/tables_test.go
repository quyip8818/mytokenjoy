package seed_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func requireDatabaseURL(t *testing.T) string {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	return dbURL
}

func truncateDomainTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			member_roles, role_permission_grants, alert_rule_notify_roles,
			budget_group_members, budget_group_departments,
			model_allowlist, key_approvals, platform_keys, provider_keys,
			operation_logs, usage_ledger, budget_consumed,
			alert_rules, model_capabilities,
			org_node_budget, budget_groups, org_nodes, members,
			roles, permissions, models,
			org_sync_logs, org_import_failures,
			org_integration, overrun_policy, audit_settings,
			companies
		RESTART IDENTITY CASCADE
	`)
	return err
}

func TestApplyTablesMatchesSnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaURL := testutil.TestSchemaURL(t)
	cfg := testutil.PreparedConfig(schemaURL)

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
		t.Fatalf("truncate domain tables: %v", err)
	}

	snap := seed.Load(cfg)
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

	assertCount(t, ctx, pool, "companies", 2)
	assertCount(t, ctx, pool, "members", len(snap.Members))
	assertCount(t, ctx, pool, "roles", len(snap.Roles))
	assertCount(t, ctx, pool, "models", len(snap.Models))
	assertCount(t, ctx, pool, "provider_keys", len(snap.ProviderKeys))
	assertCount(t, ctx, pool, "platform_keys", len(snap.PlatformKeys))
	assertCount(t, ctx, pool, "org_node_budget", len(pkgorg.FlattenOrgNodeTree(snap.OrgNodes)))
	assertSeedOrgNodeBudget(t, ctx, pool, contract.IDDept3, testutil.DisplayPoints(20000), testutil.DisplayPoints(1500))
}

func assertSeedOrgNodeBudget(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeID string, wantBudget, wantReserved float64) {
	t.Helper()
	var budget float64
	var reserved *float64
	err := pool.QueryRow(ctx, `
		SELECT budget, reserved_pool
		FROM org_node_budget
		WHERE company_id = $1 AND node_id = $2
	`, contract.DefaultCompanyID, nodeID).Scan(&budget, &reserved)
	if err != nil {
		t.Fatalf("query org_node_budget %s: %v", nodeID, err)
	}
	if budget != wantBudget {
		t.Fatalf("%s budget: got %v want %v", nodeID, budget, wantBudget)
	}
	if reserved == nil || *reserved != wantReserved {
		t.Fatalf("%s reserved_pool: got %+v want %v", nodeID, reserved, wantReserved)
	}
}

func assertCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string, want int) {
	t.Helper()
	var got int
	query := "SELECT COUNT(*) FROM " + table
	if err := pool.QueryRow(ctx, query).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s: expected %d rows, got %d", table, want, got)
	}
}
