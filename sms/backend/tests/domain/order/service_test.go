package order_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/order"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	orders []types.PurchaseOrder
}

func newMockStore() *mockStore {
	return &mockStore{}
}

func (m *mockStore) ListOrders(_ context.Context, f order.ListFilter) (*types.PagedResult[types.PurchaseOrder], error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 10
	}
	start := (page - 1) * size
	if start > len(m.orders) {
		start = len(m.orders)
	}
	end := start + size
	if end > len(m.orders) {
		end = len(m.orders)
	}
	return &types.PagedResult[types.PurchaseOrder]{
		Items: m.orders[start:end], Total: len(m.orders), Page: page, PageSize: size,
	}, nil
}

func (m *mockStore) GetOrder(_ context.Context, id uuid.UUID) (*types.PurchaseOrder, error) {
	for i := range m.orders {
		if m.orders[i].ID == id {
			return &m.orders[i], nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateOrder(_ context.Context, o *types.PurchaseOrder) error {
	o.ID = uuid.Must(uuid.NewV7())
	m.orders = append(m.orders, *o)
	return nil
}

func (m *mockStore) UpdateOrder(_ context.Context, id uuid.UUID, o *types.PurchaseOrder) error {
	for i := range m.orders {
		if m.orders[i].ID == id {
			o.ID = id
			m.orders[i] = *o
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteOrder(_ context.Context, id uuid.UUID) error {
	for i := range m.orders {
		if m.orders[i].ID == id {
			m.orders = append(m.orders[:i], m.orders[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) RecentOrders(_ context.Context, limit int) ([]types.PurchaseOrder, error) {
	if limit > len(m.orders) {
		limit = len(m.orders)
	}
	return m.orders[:limit], nil
}

// --- tests ---

var testCreatedBy = uuid.Must(uuid.NewV7())
var testSupplierID = uuid.Must(uuid.NewV7())

func newService() *order.Service {
	return order.NewService(newMockStore())
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	o, err := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-001", SupplierID: testSupplierID, Status: "pending",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.ID == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if o.OrderNo != "PO-001" {
		t.Fatalf("expected OrderNo PO-001, got %s", o.OrderNo)
	}
}

func TestCreate_DefaultStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	o, err := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-002", SupplierID: testSupplierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != "pending" {
		t.Fatalf("expected default status pending, got %s", o.Status)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testCreatedBy, order.CreateInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-X", SupplierID: testSupplierID, Status: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}
}

func TestUpdate_StatusTransition(t *testing.T) {
	t.Parallel()
	svc := newService()
	o, _ := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-T", SupplierID: testSupplierID, Status: "pending",
	})
	// pending -> approved 允许
	_, err := svc.Update(context.Background(), o.ID, order.UpdateInput{
		OrderNo: "PO-T", SupplierID: testSupplierID, Status: "approved",
	})
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}
}

func TestUpdate_InvalidTransition(t *testing.T) {
	t.Parallel()
	svc := newService()
	o, _ := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-T2", SupplierID: testSupplierID, Status: "pending",
	})
	// pending -> completed 不允许
	_, err := svc.Update(context.Background(), o.ID, order.UpdateInput{
		OrderNo: "PO-T2", SupplierID: testSupplierID, Status: "completed",
	})
	if err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestList_CapsPageSize(t *testing.T) {
	t.Parallel()
	svc := newService()
	// service caps pageSize to 100; mock receives capped value
	result, err := svc.List(context.Background(), order.ListFilter{Page: 1, PageSize: 200})
	if err != nil {
		t.Fatal(err)
	}
	if result.PageSize > 100 {
		t.Fatalf("expected pageSize capped to 100, got %d", result.PageSize)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	svc := newService()
	o, _ := svc.Create(context.Background(), testCreatedBy, order.CreateInput{
		OrderNo: "PO-DEL", SupplierID: testSupplierID, Status: "pending",
	})
	if err := svc.Delete(context.Background(), o.ID); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Get(context.Background(), o.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}
