package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	RedemptionStatusUnused   = "unused"
	RedemptionStatusUsed     = "used"
	RedemptionStatusDisabled = "disabled"

	RechargeSourceRedemption = "redemption"
)

type RedemptionCode struct {
	ID                uuid.UUID
	Code              string
	BatchName         string
	FaceValue         float64
	Currency          string
	Status            string
	RedeemedByCompany *uuid.UUID
	RedeemedByMember  *uuid.UUID
	RedeemedAt        *time.Time
	RechargeOrderID   *uuid.UUID
	ExpiresAt         time.Time
	CreatedBy         uuid.UUID
	Note              string
	CreatedAt         time.Time
}

type RedemptionListFilter struct {
	BatchName *string
	Status    *string
	Page      int
	PageSize  int
}

type RedemptionListResult struct {
	Items []RedemptionCode
	Total int
}

type RedemptionRepository interface {
	BatchInsert(ctx context.Context, codes []RedemptionCode) error
	GetCodeForUpdate(ctx context.Context, code string) (*RedemptionCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID, companyID, memberID, orderID uuid.UUID) error
	List(ctx context.Context, filter RedemptionListFilter) (RedemptionListResult, error)
}
