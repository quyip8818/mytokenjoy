package contract_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/contract"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	contracts   []types.Contract
	attachments []types.ContractAttachment
	hasOrders   map[uuid.UUID]bool
}

func newMockStore() *mockStore {
	return &mockStore{hasOrders: map[uuid.UUID]bool{}}
}

func (m *mockStore) ListContracts(_ context.Context, f contract.ListFilter) (*types.PagedResult[types.Contract], error) {
	return &types.PagedResult[types.Contract]{
		Items: m.contracts, Total: len(m.contracts), Page: f.Page, PageSize: f.PageSize,
	}, nil
}

func (m *mockStore) GetContract(_ context.Context, id uuid.UUID) (*types.ContractDetail, error) {
	for i := range m.contracts {
		if m.contracts[i].ID == id {
			return &types.ContractDetail{Contract: m.contracts[i]}, nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateContract(_ context.Context, c *types.Contract) error {
	c.ID = uuid.Must(uuid.NewV7())
	m.contracts = append(m.contracts, *c)
	return nil
}

func (m *mockStore) UpdateContract(_ context.Context, id uuid.UUID, c *types.Contract) error {
	for i := range m.contracts {
		if m.contracts[i].ID == id {
			c.ID = id
			m.contracts[i] = *c
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteContract(_ context.Context, id uuid.UUID) error {
	for i := range m.contracts {
		if m.contracts[i].ID == id {
			m.contracts = append(m.contracts[:i], m.contracts[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) HasContractOrders(_ context.Context, id uuid.UUID) (bool, error) {
	return m.hasOrders[id], nil
}

func (m *mockStore) CreateAttachment(_ context.Context, a *types.ContractAttachment) error {
	a.ID = uuid.Must(uuid.NewV7())
	m.attachments = append(m.attachments, *a)
	return nil
}

func (m *mockStore) GetAttachment(_ context.Context, id uuid.UUID) (*types.ContractAttachment, error) {
	for i := range m.attachments {
		if m.attachments[i].ID == id {
			return &m.attachments[i], nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) DeleteAttachment(_ context.Context, id uuid.UUID) error {
	for i := range m.attachments {
		if m.attachments[i].ID == id {
			m.attachments = append(m.attachments[:i], m.attachments[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) ListAttachments(_ context.Context, contractID uuid.UUID) ([]types.ContractAttachment, error) {
	var out []types.ContractAttachment
	for _, a := range m.attachments {
		if a.ContractID == contractID {
			out = append(out, a)
		}
	}
	return out, nil
}

// --- tests ---

var testCreatedBy = uuid.Must(uuid.NewV7())
var testSupplierID = uuid.Must(uuid.NewV7())

func newService() *contract.Service {
	return contract.NewService(newMockStore(), "/tmp/test-uploads")
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	c, err := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "CT-001", Title: "服务合同",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if c.Status != "draft" {
		t.Fatalf("expected default status draft, got %s", c.Status)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := newService()
	cases := []struct {
		name  string
		input contract.CreateInput
	}{
		{"empty title", contract.CreateInput{SupplierID: testSupplierID, ContractNo: "X"}},
		{"empty contractNo", contract.CreateInput{SupplierID: testSupplierID, Title: "X"}},
		{"empty supplier", contract.CreateInput{ContractNo: "X", Title: "X"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), testCreatedBy, tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "X", Title: "X", Status: "bad",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "CT-U", Title: "Original",
	})
	updated, err := svc.Update(context.Background(), created.ID, contract.UpdateInput{
		SupplierID: testSupplierID, ContractNo: "CT-U", Title: "Updated", Status: "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected Updated, got %s", updated.Title)
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "CT-S", Title: "X",
	})
	_, err := svc.Update(context.Background(), created.ID, contract.UpdateInput{
		SupplierID: testSupplierID, ContractNo: "CT-S", Title: "X", Status: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "CT-D", Title: "Del",
	})
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
}

func TestDelete_WithOrders(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	svc := contract.NewService(store, "/tmp/test-uploads")
	created, _ := svc.Create(context.Background(), testCreatedBy, contract.CreateInput{
		SupplierID: testSupplierID, ContractNo: "CT-REF", Title: "HasOrders",
	})
	store.hasOrders[created.ID] = true
	err := svc.Delete(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error when contract has orders")
	}
}

func TestList_CapsPageSize(t *testing.T) {
	t.Parallel()
	svc := newService()
	result, err := svc.List(context.Background(), contract.ListFilter{Page: 1, PageSize: 200})
	if err != nil {
		t.Fatal(err)
	}
	if result.PageSize > 100 {
		t.Fatalf("expected pageSize capped to 100, got %d", result.PageSize)
	}
}
