package contract

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListContracts(ctx context.Context, f ListFilter) (*types.PagedResult[types.Contract], error)
	GetContract(ctx context.Context, id uuid.UUID) (*types.ContractDetail, error)
	CreateContract(ctx context.Context, c *types.Contract) error
	UpdateContract(ctx context.Context, id uuid.UUID, c *types.Contract) error
	DeleteContract(ctx context.Context, id uuid.UUID) error
	HasContractOrders(ctx context.Context, id uuid.UUID) (bool, error)
	CreateAttachment(ctx context.Context, a *types.ContractAttachment) error
	GetAttachment(ctx context.Context, id uuid.UUID) (*types.ContractAttachment, error)
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	ListAttachments(ctx context.Context, contractID uuid.UUID) ([]types.ContractAttachment, error)
}

type ListFilter struct {
	Page       int
	PageSize   int
	Keyword    string
	SupplierID uuid.UUID
	Status     string
}

type CreateInput struct {
	SupplierID uuid.UUID `json:"supplierId"`
	ContractNo string    `json:"contractNo"`
	Title      string    `json:"title"`
	Amount     *float64  `json:"amount"`
	SignDate   *string   `json:"signDate"`
	StartDate  *string   `json:"startDate"`
	EndDate    *string   `json:"endDate"`
	Status     string    `json:"status"`
	Remarks    *string   `json:"remarks"`
}

type UpdateInput = CreateInput

type Service struct {
	store     Store
	uploadDir string
}

func NewService(store Store, uploadDir string) *Service {
	return &Service{store: store, uploadDir: uploadDir}
}

func (s *Service) List(ctx context.Context, f ListFilter) (*types.PagedResult[types.Contract], error) {
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return s.store.ListContracts(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*types.ContractDetail, error) {
	return s.store.GetContract(ctx, id)
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

func (s *Service) Create(ctx context.Context, createdBy uuid.UUID, input CreateInput) (*types.Contract, error) {
	if input.Title == "" || input.ContractNo == "" || input.SupplierID == uuid.Nil {
		return nil, fmt.Errorf("%w: 合同标题、编号和供应商不能为空", types.ErrValidation)
	}
	if input.Status == "" {
		input.Status = "draft"
	}
	if !types.IsValidStatus(input.Status, types.ContractStatuses) {
		return nil, fmt.Errorf("%w: 无效的合同状态", types.ErrValidation)
	}
	c := &types.Contract{
		SupplierID: input.SupplierID,
		ContractNo: input.ContractNo,
		Title:      input.Title,
		Amount:     input.Amount,
		SignDate:   parseDatePtr(input.SignDate),
		StartDate:  parseDatePtr(input.StartDate),
		EndDate:    parseDatePtr(input.EndDate),
		Status:     input.Status,
		Remarks:    input.Remarks,
		CreatedBy:  &createdBy,
	}
	if err := s.store.CreateContract(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.Contract, error) {
	if input.Title == "" || input.ContractNo == "" || input.SupplierID == uuid.Nil {
		return nil, fmt.Errorf("%w: 合同标题、编号和供应商不能为空", types.ErrValidation)
	}
	if input.Status != "" && !types.IsValidStatus(input.Status, types.ContractStatuses) {
		return nil, fmt.Errorf("%w: 无效的合同状态", types.ErrValidation)
	}
	c := &types.Contract{
		SupplierID: input.SupplierID,
		ContractNo: input.ContractNo,
		Title:      input.Title,
		Amount:     input.Amount,
		SignDate:   parseDatePtr(input.SignDate),
		StartDate:  parseDatePtr(input.StartDate),
		EndDate:    parseDatePtr(input.EndDate),
		Status:     input.Status,
		Remarks:    input.Remarks,
	}
	if err := s.store.UpdateContract(ctx, id, c); err != nil {
		return nil, err
	}
	c.ID = id
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	hasOrders, err := s.store.HasContractOrders(ctx, id)
	if err != nil {
		return err
	}
	if hasOrders {
		return fmt.Errorf("%w: 该合同有关联的订单，无法删除", types.ErrHasRefs)
	}
	return s.store.DeleteContract(ctx, id)
}

func (s *Service) UploadAttachment(ctx context.Context, contractID uuid.UUID, uploadedBy uuid.UUID, header *multipart.FileHeader, file multipart.File) (*types.ContractAttachment, error) {
	_ = os.MkdirAll(s.uploadDir, 0o755)
	ext := filepath.Ext(header.Filename)
	storedName := uuid.NewString() + ext
	dstPath := filepath.Join(s.uploadDir, storedName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	att := &types.ContractAttachment{
		ContractID: contractID,
		FileName:   header.Filename,
		FilePath:   dstPath,
		FileSize:   header.Size,
		UploadedBy: &uploadedBy,
	}
	if err := s.store.CreateAttachment(ctx, att); err != nil {
		_ = os.Remove(dstPath)
		return nil, err
	}
	return att, nil
}

func (s *Service) GetAttachment(ctx context.Context, id uuid.UUID) (*types.ContractAttachment, error) {
	return s.store.GetAttachment(ctx, id)
}

func (s *Service) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	att, err := s.store.GetAttachment(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAttachment(ctx, id); err != nil {
		return err
	}
	_ = os.Remove(att.FilePath)
	return nil
}
