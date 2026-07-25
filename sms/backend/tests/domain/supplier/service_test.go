package supplier_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/supplier"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	suppliers []types.Supplier
	contacts  []types.SupplierContact
	hasRefs   map[uuid.UUID]bool
}

func newMockStore() *mockStore {
	return &mockStore{hasRefs: map[uuid.UUID]bool{}}
}

func (m *mockStore) ListSuppliers(_ context.Context, f supplier.ListFilter) (*types.PagedResult[types.Supplier], error) {
	return &types.PagedResult[types.Supplier]{
		Items: m.suppliers, Total: len(m.suppliers), Page: f.Page, PageSize: f.PageSize,
	}, nil
}

func (m *mockStore) GetSupplier(_ context.Context, id uuid.UUID) (*types.SupplierDetail, error) {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			return &types.SupplierDetail{Supplier: m.suppliers[i]}, nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateSupplier(_ context.Context, s *types.Supplier) error {
	s.ID = uuid.Must(uuid.NewV7())
	m.suppliers = append(m.suppliers, *s)
	return nil
}

func (m *mockStore) UpdateSupplier(_ context.Context, id uuid.UUID, s *types.Supplier) error {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			s.ID = id
			m.suppliers[i] = *s
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteSupplier(_ context.Context, id uuid.UUID) error {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			m.suppliers = append(m.suppliers[:i], m.suppliers[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) HasSupplierRefs(_ context.Context, id uuid.UUID) (bool, error) {
	return m.hasRefs[id], nil
}

func (m *mockStore) CreateContact(_ context.Context, c *types.SupplierContact) error {
	c.ID = uuid.Must(uuid.NewV7())
	m.contacts = append(m.contacts, *c)
	return nil
}

func (m *mockStore) UpdateContact(_ context.Context, c *types.SupplierContact) error {
	for i := range m.contacts {
		if m.contacts[i].ID == c.ID {
			m.contacts[i] = *c
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteContact(_ context.Context, id uuid.UUID) error {
	for i := range m.contacts {
		if m.contacts[i].ID == id {
			m.contacts = append(m.contacts[:i], m.contacts[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) SupplierOptions(_ context.Context) ([]types.IDName, error) {
	opts := make([]types.IDName, len(m.suppliers))
	for i, s := range m.suppliers {
		opts[i] = types.IDName{ID: s.ID, Name: s.Name}
	}
	return opts, nil
}

// --- tests ---

var testCreatedBy = uuid.Must(uuid.NewV7())

func newService() *supplier.Service {
	return supplier.NewService(newMockStore())
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	s, err := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{
		Name: "Acme", Code: "ACM", Status: "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.ID == uuid.Nil || s.Name != "Acme" {
		t.Fatalf("unexpected supplier %+v", s)
	}
}

func TestCreate_DefaultStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	s, err := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{
		Name: "Test", Code: "TST",
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != "potential" {
		t.Fatalf("expected default status potential, got %s", s.Status)
	}
}

func TestCreate_EmptyName(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{Code: "X"})
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{
		Name: "X", Code: "X", Status: "bad",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}
}

func TestDelete_WithRefs(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	svc := supplier.NewService(store)
	s, _ := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{
		Name: "Ref", Code: "REF", Status: "active",
	})
	store.hasRefs[s.ID] = true
	err := svc.Delete(context.Background(), s.ID)
	if err == nil {
		t.Fatal("expected error when supplier has refs")
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	s, _ := svc.Create(context.Background(), testCreatedBy, supplier.CreateInput{
		Name: "Del", Code: "DEL", Status: "active",
	})
	if err := svc.Delete(context.Background(), s.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCreateContact_EmptyName(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.CreateContact(context.Background(), uuid.Must(uuid.NewV7()), supplier.ContactInput{})
	if err == nil {
		t.Fatal("expected validation error for empty contact name")
	}
}

func TestCreateContact_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	c, err := svc.CreateContact(context.Background(), uuid.Must(uuid.NewV7()), supplier.ContactInput{Name: "张三"})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID == uuid.Nil {
		t.Fatal("expected non-nil contact ID")
	}
}
