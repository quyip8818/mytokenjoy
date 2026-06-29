package budget_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpdateNodeSuccess(t *testing.T) {
	svc, st := newBudgetService(t)
	reserved := 1500.0
	updated, err := svc.UpdateNode(context.Background(), seed.IDDept3, 21000, &reserved)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Budget != 21000 {
		t.Fatalf("expected budget 21000, got %v", updated.Budget)
	}
	node := findDept3(st.Budget().Tree())
	if node == nil || node.Budget != 21000 {
		t.Fatalf("expected persisted budget 21000, got %+v", node)
	}
}

func TestUpdateNodeOversell(t *testing.T) {
	svc, _ := newBudgetService(t)
	reserved := 1500.0
	_, err := svc.UpdateNode(context.Background(), seed.IDDept3, 90000, &reserved)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberQuotaBelowAllocated(t *testing.T) {
	svc, _ := newBudgetService(t)
	_, err := svc.UpdateMemberQuota(context.Background(), seed.IDMember1, 1000)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberQuotaSuccess(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	snapshot := seed.Load(cfg)
	members := make([]types.Member, 0, len(snapshot.Members))
	for _, member := range snapshot.Members {
		if member.DepartmentID == seed.IDDept3 && member.ID != seed.IDMember1 {
			continue
		}
		members = append(members, member)
	}
	snapshot.Members = members
	st = store.NewMemory(snapshot)
	svc := budget.NewService(cfg, st, simulate.NewDelayer(false))

	result, err := svc.UpdateMemberQuota(context.Background(), seed.IDMember1, 15000)
	if err != nil {
		t.Fatal(err)
	}
	if result.PersonalQuota != 15000 {
		t.Fatalf("expected personal quota 15000, got %v", result.PersonalQuota)
	}
	pool := st.Budget().MemberQuotaPools()[seed.IDMember1]
	if pool.PersonalQuota != 15000 {
		t.Fatalf("expected pool personal quota 15000, got %v", pool.PersonalQuota)
	}
}

func TestListMemberQuotasUnknownDept(t *testing.T) {
	svc, _ := newBudgetService(t)
	_, err := svc.ListMemberQuotas("dept-missing")
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestCreateGroup(t *testing.T) {
	svc, st := newBudgetService(t)
	created, err := svc.CreateGroup(context.Background(), types.BudgetGroup{
		Name: "Test Group", Budget: 5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected created group id")
	}
	found := false
	for _, group := range st.Budget().Groups() {
		if group.ID == created.ID && group.Name == "Test Group" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created group not found in store")
	}
}

func TestDeleteGroup(t *testing.T) {
	svc, st := newBudgetService(t)
	if err := svc.DeleteGroup(seed.IDBudgetGroup4); err != nil {
		t.Fatal(err)
	}
	for _, group := range st.Budget().Groups() {
		if group.ID == seed.IDBudgetGroup4 {
			t.Fatal("expected bg-4 deleted")
		}
	}
}
