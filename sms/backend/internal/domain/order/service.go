package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListOrders(ctx context.Context, f ListFilter) (*types.PagedResult[types.PurchaseOrder], error)
	GetOrder(ctx context.Context, id uuid.UUID) (*types.PurchaseOrder, error)
	CreateOrder(ctx context.Context, o *types.PurchaseOrder) error
	UpdateOrder(ctx context.Context, id uuid.UUID, o *types.PurchaseOrder) error
	DeleteOrder(ctx context.Context, id uuid.UUID) error
	RecentOrders(ctx context.Context, limit int) ([]types.PurchaseOrder, error)
}

type ListFilter struct {
	Page       int
	PageSize   int
	Keyword    string
	SupplierID uuid.UUID
	Status     string
}

type CreateInput struct {
	OrderNo     string     `json:"orderNo"`
	SupplierID  uuid.UUID  `json:"supplierId"`
	ContractID  *uuid.UUID `json:"contractId"`
	TotalAmount *float64   `json:"totalAmount"`
	OrderDate   *string    `json:"orderDate"`
	Status      string     `json:"status"`
	Description *string    `json:"description"`
}

type UpdateInput struct {
	OrderNo     string     `json:"orderNo"`
	SupplierID  uuid.UUID  `json:"supplierId"`
	ContractID  *uuid.UUID `json:"contractId"`
	TotalAmount *float64   `json:"totalAmount"`
	OrderDate   *string    `json:"orderDate"`
	Status      string     `json:"status"`
	Description *string    `json:"description"`
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, f ListFilter) (*types.PagedResult[types.PurchaseOrder], error) {
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return s.store.ListOrders(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*types.PurchaseOrder, error) {
	return s.store.GetOrder(ctx, id)
}

func parseDatePtr(s *string) *types.DateOnly {
	if s == nil || *s == "" {
		return nil
	}
	d, err := types.ParseDateOnly(*s)
	if err != nil {
		return nil
	}
	return &d
}

func (s *Service) Create(ctx context.Context, createdBy uuid.UUID, input CreateInput) (*types.PurchaseOrder, error) {
	if input.OrderNo == "" || input.SupplierID == uuid.Nil {
		return nil, fmt.Errorf("%w: 订单号和供应商不能为空", types.ErrValidation)
	}
	if input.Status == "" {
		input.Status = "pending"
	}
	if !types.IsValidStatus(input.Status, types.OrderStatuses) {
		return nil, fmt.Errorf("%w: 无效的订单状态", types.ErrValidation)
	}
	o := &types.PurchaseOrder{
		OrderNo:     input.OrderNo,
		SupplierID:  input.SupplierID,
		ContractID:  input.ContractID,
		TotalAmount: input.TotalAmount,
		OrderDate:   parseDatePtr(input.OrderDate),
		Status:      input.Status,
		Description: input.Description,
		CreatedBy:   &createdBy,
	}
	if err := s.store.CreateOrder(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.PurchaseOrder, error) {
	if input.Status != "" {
		existing, err := s.store.GetOrder(ctx, id)
		if err != nil {
			return nil, err
		}
		if existing.Status != input.Status && !types.IsValidTransition(existing.Status, input.Status) {
			return nil, fmt.Errorf("%w: 不允许从 %s 转为 %s", types.ErrValidation, existing.Status, input.Status)
		}
	}
	o := &types.PurchaseOrder{
		OrderNo:     input.OrderNo,
		SupplierID:  input.SupplierID,
		ContractID:  input.ContractID,
		TotalAmount: input.TotalAmount,
		OrderDate:   parseDatePtr(input.OrderDate),
		Status:      input.Status,
		Description: input.Description,
	}
	if err := s.store.UpdateOrder(ctx, id, o); err != nil {
		return nil, err
	}
	o.ID = id
	return o, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteOrder(ctx, id)
}
