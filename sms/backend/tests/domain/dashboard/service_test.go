package dashboard_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/dashboard"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	cards     *types.DashboardCards
	charts    *types.DashboardCharts
	expiring  []types.ExpiringContract
	orders    []types.PurchaseOrder
	lastDays  int
	lastLimit int
}

func newMockStore() *mockStore {
	return &mockStore{
		cards: &types.DashboardCards{
			SupplierTotal: 10, ActiveSuppliers: 8, ModelTotal: 20, ActiveContracts: 5,
		},
		charts: &types.DashboardCharts{
			GradeDistribution:    []types.LabelCount{{Label: "A", Count: 3}},
			ModelCountBySupplier: []types.LabelCount{{Label: "Acme", Count: 5}},
		},
		expiring: []types.ExpiringContract{
			{ID: uuid.Must(uuid.NewV7()), Title: "即将到期", ContractNo: "CT-1", EndDate: "2024-02-01", SupplierName: "Acme"},
		},
		orders: []types.PurchaseOrder{
			{ID: uuid.Must(uuid.NewV7()), OrderNo: "PO-1", Status: "pending"},
			{ID: uuid.Must(uuid.NewV7()), OrderNo: "PO-2", Status: "approved"},
		},
	}
}

func (m *mockStore) DashboardCards(_ context.Context) (*types.DashboardCards, error) {
	return m.cards, nil
}

func (m *mockStore) DashboardCharts(_ context.Context) (*types.DashboardCharts, error) {
	return m.charts, nil
}

func (m *mockStore) ExpiringContracts(_ context.Context, days int) ([]types.ExpiringContract, error) {
	m.lastDays = days
	return m.expiring, nil
}

func (m *mockStore) RecentOrders(_ context.Context, limit int) ([]types.PurchaseOrder, error) {
	m.lastLimit = limit
	if limit > len(m.orders) {
		limit = len(m.orders)
	}
	return m.orders[:limit], nil
}

// --- tests ---

func TestCards(t *testing.T) {
	t.Parallel()
	svc := dashboard.NewService(newMockStore())
	cards, err := svc.Cards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cards.SupplierTotal != 10 {
		t.Fatalf("expected 10, got %d", cards.SupplierTotal)
	}
	if cards.ActiveContracts != 5 {
		t.Fatalf("expected 5, got %d", cards.ActiveContracts)
	}
}

func TestCharts(t *testing.T) {
	t.Parallel()
	svc := dashboard.NewService(newMockStore())
	charts, err := svc.Charts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(charts.GradeDistribution) != 1 {
		t.Fatalf("expected 1 grade entry, got %d", len(charts.GradeDistribution))
	}
}

func TestExpiring_Uses30Days(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	svc := dashboard.NewService(store)
	contracts, err := svc.Expiring(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if store.lastDays != 30 {
		t.Fatalf("expected 30 days, got %d", store.lastDays)
	}
	if len(contracts) != 1 {
		t.Fatalf("expected 1 expiring contract, got %d", len(contracts))
	}
}

func TestRecentOrders_Uses5Limit(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	svc := dashboard.NewService(store)
	orders, err := svc.RecentOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if store.lastLimit != 5 {
		t.Fatalf("expected limit 5, got %d", store.lastLimit)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders (all available), got %d", len(orders))
	}
}
