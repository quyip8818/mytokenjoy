package model

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListModels(ctx context.Context, f ListFilter) (*types.PagedResult[types.AiModel], error)
	ListModelsWithModelID(ctx context.Context) ([]types.AiModel, error)
	GetModel(ctx context.Context, id uuid.UUID) (*types.AiModel, error)
	CreateModel(ctx context.Context, m *types.AiModel) error
	UpdateModel(ctx context.Context, id uuid.UUID, m *types.AiModel) error
	DeleteModel(ctx context.Context, id uuid.UUID) error
}

type ListFilter struct {
	Page       int
	PageSize   int
	Keyword    string
	SupplierID *uuid.UUID
	ModelType  string
	Status     string
}

type CreateInput struct {
	SupplierID    *uuid.UUID `json:"supplierId"`
	ModelName     string    `json:"modelName"`
	ModelID       *string   `json:"modelId"`
	ModelType     *string   `json:"modelType"`
	ContextLength *int      `json:"contextLength"`
	InputPrice    *float64  `json:"inputPrice"`
	OutputPrice   *float64  `json:"outputPrice"`
	Discount      *float64  `json:"discount"`
	Status        string    `json:"status"`
	Description   *string   `json:"description"`
}

type UpdateInput = CreateInput

// Syncer 是可选的同步回调接口，由 newapisync.Service 实现
type Syncer interface {
	UpsertOne(ctx context.Context, modelID string, inputPrice, outputPrice float64)
}

type Service struct {
	store  Store
	syncer Syncer
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// SetSyncer 注入同步回调（app 层在组装后调用）
func (s *Service) SetSyncer(syncer Syncer) {
	s.syncer = syncer
}

func (s *Service) List(ctx context.Context, f ListFilter) (*types.PagedResult[types.AiModel], error) {
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return s.store.ListModels(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*types.AiModel, error) {
	return s.store.GetModel(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*types.AiModel, error) {
	if input.ModelName == "" {
		return nil, fmt.Errorf("%w: 模型名称和供应商不能为空", types.ErrValidation)
	}
	if input.Status == "" {
		input.Status = "available"
	}
	if !types.IsValidStatus(input.Status, types.ModelStatuses) {
		return nil, fmt.Errorf("%w: 无效的模型状态", types.ErrValidation)
	}
	m := &types.AiModel{
		SupplierID:    input.SupplierID,
		ModelName:     input.ModelName,
		ModelID:       input.ModelID,
		ModelType:     input.ModelType,
		ContextLength: input.ContextLength,
		InputPrice:    input.InputPrice,
		OutputPrice:   input.OutputPrice,
		Discount:      input.Discount,
		Status:        input.Status,
		Description:   input.Description,
	}
	if err := s.store.CreateModel(ctx, m); err != nil {
		return nil, err
	}
	s.triggerSync(m)
	return m, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.AiModel, error) {
	if input.ModelName == "" {
		return nil, fmt.Errorf("%w: 模型名称和供应商不能为空", types.ErrValidation)
	}
	if input.Status != "" && !types.IsValidStatus(input.Status, types.ModelStatuses) {
		return nil, fmt.Errorf("%w: 无效的模型状态", types.ErrValidation)
	}
	m := &types.AiModel{
		SupplierID:    input.SupplierID,
		ModelName:     input.ModelName,
		ModelID:       input.ModelID,
		ModelType:     input.ModelType,
		ContextLength: input.ContextLength,
		InputPrice:    input.InputPrice,
		OutputPrice:   input.OutputPrice,
		Discount:      input.Discount,
		Status:        input.Status,
		Description:   input.Description,
	}
	if err := s.store.UpdateModel(ctx, id, m); err != nil {
		return nil, err
	}
	m.ID = id
	s.triggerSync(m)
	return m, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteModel(ctx, id)
}

func (s *Service) ListWithModelID(ctx context.Context) ([]types.AiModel, error) {
	return s.store.ListModelsWithModelID(ctx)
}

// triggerSync 异步推送价格到 NewAPI（fire-and-forget）
func (s *Service) triggerSync(m *types.AiModel) {
	if s.syncer == nil || m.ModelID == nil || *m.ModelID == "" {
		return
	}
	inp := 0.0
	out := 0.0
	if m.InputPrice != nil {
		inp = *m.InputPrice
	}
	if m.OutputPrice != nil {
		out = *m.OutputPrice
	}
	if inp <= 0 {
		return
	}
	modelID := *m.ModelID
	go s.syncer.UpsertOne(context.Background(), modelID, inp, out)
}
