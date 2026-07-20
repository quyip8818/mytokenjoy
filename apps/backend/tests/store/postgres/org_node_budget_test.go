package postgres_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestOrgNodeBudgetRepositoryRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	repo := st.Budget().OrgNodeBudget()

	reserved := budgetfix.QuotaFromDisplay(800)
	row := store.OrgNodeBudgetRow{
		NodeID: contract.IDDept3, Budget: 900, ReservedPool: &reserved,
		Period: pkgbudget.PeriodMonthly, MemberAvgBudget: 120,
	}
	if err := repo.Upsert(ctx, contract.IDDept3, row); err != nil {
		t.Fatal(err)
	}
	got, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("get budget: found=%v err=%v", found, err)
	}
	if got.Budget != 900 || got.MemberAvgBudget != 120 {
		t.Fatalf("unexpected row: %+v", got)
	}
	if got.ReservedPool == nil || *got.ReservedPool != reserved {
		t.Fatalf("reserved pool mismatch: %+v", got.ReservedPool)
	}
	if got.Period != pkgbudget.PeriodMonthly {
		t.Fatalf("period mismatch: %q", got.Period)
	}
}

func TestPersistNodeBudgetPreservesPeriodAndMemberAvg(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	repo := st.Budget().OrgNodeBudget()

	before, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("seed budget missing: found=%v err=%v", found, err)
	}
	newBudget := budgetfix.QuotaFromDisplay(22000)
	if err := pkgbudget.PersistNodeBudget(ctx, repo, contract.IDDept3, types.BudgetNode{
		Budget: newBudget, ReservedPool: before.ReservedPool,
	}); err != nil {
		t.Fatal(err)
	}
	got, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("get budget: found=%v err=%v", found, err)
	}
	if got.Budget != newBudget {
		t.Fatalf("budget: got %v want %v", got.Budget, newBudget)
	}
	if got.Period != before.Period {
		t.Fatalf("period changed: got %q want %q", got.Period, before.Period)
	}
	if got.MemberAvgBudget != before.MemberAvgBudget {
		t.Fatalf("member avg changed: got %v want %v", got.MemberAvgBudget, before.MemberAvgBudget)
	}
}

func TestPersistMemberAvgBudgetUpdatesOnlyMemberAvg(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	repo := st.Budget().OrgNodeBudget()

	before, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("seed budget missing: found=%v err=%v", found, err)
	}
	wantAvg := budgetfix.QuotaFromDisplay(16000)
	if err := pkgbudget.PersistMemberAvgBudget(ctx, repo, contract.IDDept3, wantAvg); err != nil {
		t.Fatal(err)
	}
	got, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("get budget: found=%v err=%v", found, err)
	}
	if got.MemberAvgBudget != wantAvg {
		t.Fatalf("member avg: got %v want %v", got.MemberAvgBudget, wantAvg)
	}
	if got.Budget != before.Budget {
		t.Fatalf("budget changed: got %v want %v", got.Budget, before.Budget)
	}
}

func TestSetTreeDoesNotOverwriteOrgNodeBudget(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	repo := st.Budget().OrgNodeBudget()

	wantBudget := budgetfix.QuotaFromDisplay(20500)
	reserved := budgetfix.QuotaFromDisplay(1500)
	if err := repo.Upsert(ctx, contract.IDDept3, store.OrgNodeBudgetRow{
		NodeID: contract.IDDept3, Budget: wantBudget, ReservedPool: &reserved,
		Period: pkgbudget.PeriodMonthly,
	}); err != nil {
		t.Fatal(err)
	}

	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var touch func([]types.OrgNode)
	touch = func(list []types.OrgNode) {
		for i := range list {
			if list[i].ID == contract.IDDept3 {
				list[i].Name = list[i].Name + " StoreTest"
				list[i].Budget = 0
				list[i].ReservedPool = nil
			}
			if len(list[i].Children) > 0 {
				touch(list[i].Children)
			}
		}
	}
	touch(nodes)
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}

	got, found, err := repo.Get(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("get budget: found=%v err=%v", found, err)
	}
	if got.Budget != wantBudget {
		t.Fatalf("budget overwritten by SetTree: got %v want %v", got.Budget, wantBudget)
	}
	if got.ReservedPool == nil || *got.ReservedPool != reserved {
		t.Fatalf("reserved overwritten by SetTree: %+v", got.ReservedPool)
	}
}
