package redemption

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	Generate(ctx context.Context, input GenerateInput) (GenerateResult, error)
	List(ctx context.Context, filter store.RedemptionListFilter) (store.RedemptionListResult, error)
}

type GenerateInput struct {
	BatchName     string
	FaceValue     float64
	Currency      string
	Quantity      int
	ExpiresInDays int
	Note          string
	CreatedBy     uuid.UUID
}

type GenerateResult struct {
	BatchName string
	FaceValue float64
	Quantity  int
	ExpiresAt time.Time
}

// Store is the narrow store surface needed for redemption code management.
type Store interface {
	Redemption() store.RedemptionRepository
}

type service struct {
	store Store
}

func NewService(st Store) Service {
	return &service{store: st}
}

func (s *service) Generate(ctx context.Context, input GenerateInput) (GenerateResult, error) {
	if input.FaceValue <= 0 {
		return GenerateResult{}, domain.Validation("面值必须大于 0")
	}
	if input.Quantity <= 0 || input.Quantity > 1000 {
		return GenerateResult{}, domain.Validation("数量必须在 1-1000 之间")
	}
	if input.ExpiresInDays < 1 || input.ExpiresInDays > 1095 {
		return GenerateResult{}, domain.Validation("有效天数必须在 1-1095 之间")
	}

	currency := input.Currency
	if currency == "" {
		currency = "CNY"
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, input.ExpiresInDays)

	codes := make([]store.RedemptionCode, 0, input.Quantity)
	for i := 0; i < input.Quantity; i++ {
		code, err := GenerateCode()
		if err != nil {
			return GenerateResult{}, fmt.Errorf("generate code: %w", err)
		}
		codes = append(codes, store.RedemptionCode{
			ID:        uuid.Must(uuid.NewV7()),
			Code:      code,
			BatchName: input.BatchName,
			FaceValue: input.FaceValue,
			Currency:  currency,
			Status:    store.RedemptionStatusUnused,
			ExpiresAt: expiresAt,
			CreatedBy: input.CreatedBy,
			Note:      input.Note,
			CreatedAt: now,
		})
	}

	if err := s.store.Redemption().BatchInsert(ctx, codes); err != nil {
		return GenerateResult{}, fmt.Errorf("batch insert: %w", err)
	}

	return GenerateResult{
		BatchName: input.BatchName,
		FaceValue: input.FaceValue,
		Quantity:  input.Quantity,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *service) List(ctx context.Context, filter store.RedemptionListFilter) (store.RedemptionListResult, error) {
	return s.store.Redemption().List(ctx, filter)
}
