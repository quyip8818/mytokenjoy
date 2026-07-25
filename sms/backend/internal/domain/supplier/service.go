package supplier

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListSuppliers(ctx context.Context, f ListFilter) (*types.PagedResult[types.Supplier], error)
	GetSupplier(ctx context.Context, id uuid.UUID) (*types.SupplierDetail, error)
	CreateSupplier(ctx context.Context, s *types.Supplier) error
	UpdateSupplier(ctx context.Context, id uuid.UUID, s *types.Supplier) error
	DeleteSupplier(ctx context.Context, id uuid.UUID) error
	HasSupplierRefs(ctx context.Context, id uuid.UUID) (bool, error)
	CreateContact(ctx context.Context, c *types.SupplierContact) error
	UpdateContact(ctx context.Context, c *types.SupplierContact) error
	DeleteContact(ctx context.Context, id uuid.UUID) error
	SupplierOptions(ctx context.Context) ([]types.IDName, error)
}

type ListFilter struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	Category string
}

type CreateInput struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Category    *string `json:"category"`
	Website     *string `json:"website"`
	Status      string  `json:"status"`
	Description *string `json:"description"`
}

type UpdateInput struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Category    *string `json:"category"`
	Website     *string `json:"website"`
	Status      string  `json:"status"`
	Description *string `json:"description"`
}

type ContactInput struct {
	Name      string  `json:"name"`
	Position  *string `json:"position"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	IsPrimary bool    `json:"isPrimary"`
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, f ListFilter) (*types.PagedResult[types.Supplier], error) {
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return s.store.ListSuppliers(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*types.SupplierDetail, error) {
	return s.store.GetSupplier(ctx, id)
}

func (s *Service) Create(ctx context.Context, createdBy uuid.UUID, input CreateInput) (*types.Supplier, error) {
	if input.Name == "" || input.Code == "" {
		return nil, fmt.Errorf("%w: 厂商名称和编码不能为空", types.ErrValidation)
	}
	if input.Status == "" {
		input.Status = "potential"
	}
	if !types.IsValidStatus(input.Status, types.SupplierStatuses) {
		return nil, fmt.Errorf("%w: 无效的供应商状态", types.ErrValidation)
	}
	supplier := &types.Supplier{
		Name: input.Name, Code: input.Code,
		Category: input.Category, Website: input.Website,
		Status: input.Status, Description: input.Description,
		CreatedBy: &createdBy,
	}
	if err := s.store.CreateSupplier(ctx, supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.Supplier, error) {
	if input.Name == "" || input.Code == "" {
		return nil, fmt.Errorf("%w: 厂商名称和编码不能为空", types.ErrValidation)
	}
	if input.Status != "" && !types.IsValidStatus(input.Status, types.SupplierStatuses) {
		return nil, fmt.Errorf("%w: 无效的供应商状态", types.ErrValidation)
	}
	supplier := &types.Supplier{
		Name: input.Name, Code: input.Code,
		Category: input.Category, Website: input.Website,
		Status: input.Status, Description: input.Description,
	}
	if err := s.store.UpdateSupplier(ctx, id, supplier); err != nil {
		return nil, err
	}
	supplier.ID = id
	return supplier, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	hasRefs, err := s.store.HasSupplierRefs(ctx, id)
	if err != nil {
		return err
	}
	if hasRefs {
		return fmt.Errorf("%w: 该供应商有关联的合同或订单，无法删除", types.ErrHasRefs)
	}
	return s.store.DeleteSupplier(ctx, id)
}

func (s *Service) Options(ctx context.Context) ([]types.IDName, error) {
	return s.store.SupplierOptions(ctx)
}

func (s *Service) CreateContact(ctx context.Context, supplierID uuid.UUID, input ContactInput) (*types.SupplierContact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: 联系人姓名不能为空", types.ErrValidation)
	}
	c := &types.SupplierContact{
		SupplierID: supplierID,
		Name:       input.Name,
		Position:   input.Position,
		Phone:      input.Phone,
		Email:      input.Email,
		IsPrimary:  input.IsPrimary,
	}
	if err := s.store.CreateContact(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateContact(ctx context.Context, id uuid.UUID, input ContactInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: 联系人姓名不能为空", types.ErrValidation)
	}
	c := &types.SupplierContact{
		ID:        id,
		Name:      input.Name,
		Position:  input.Position,
		Phone:     input.Phone,
		Email:     input.Email,
		IsPrimary: input.IsPrimary,
	}
	return s.store.UpdateContact(ctx, c)
}

func (s *Service) DeleteContact(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteContact(ctx, id)
}
