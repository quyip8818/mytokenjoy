package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	RechargeStatusPending   = "pending"
	RechargeStatusConfirmed = "confirmed"
	RechargeStatusFailed    = "failed"

	RechargeSourcePlatform = "platform"
	RechargeSourceSelf     = "self"
	RechargeSourceGift     = "gift"
	RechargeSourceAdjust   = "adjust"
	RechargeSourceSystem   = "system"

	LotKindPaid      = "paid"
	LotKindGift      = "gift"
	LotKindAdjust    = "adjust"
	LotKindOverdraft = "overdraft"
	LotKindMock      = "mock"

	LotStatusActive    = "active"
	LotStatusExhausted = "exhausted"

	ActorTypeMember   = "member"
	ActorTypePlatform = "platform"

	InvoiceStatusNone    = "none"
	InvoiceStatusApplied = "applied"
	InvoiceStatusIssued  = "issued"

	PaymentMethodAlipay = "alipay"
	PaymentMethodWechat = "wechat"
)

type RechargeOrder struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	Amount         float64
	Currency       string
	PointsPerUnit  int64
	PointsGranted  float64
	Source         string
	LotKind        string
	IdempotencyKey *string
	Status         string
	DisplayOrderID string
	PaymentMethod  string
	InvoiceStatus  string
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RechargeLot struct {
	ID               uuid.UUID
	CompanyID        uuid.UUID
	RechargeOrderID  uuid.UUID
	BillingCurrency  string
	LotKind          string
	AmountDisplay    float64
	PointsGranted    float64
	PointsRemaining  float64
	UnitPriceDisplay float64
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type WalletCurrencyBalance struct {
	Currency      string
	Balance       float64
	TotalTopup    float64
	TotalConsumed float64
}

type WalletAggregate struct {
	BillingCurrency string
	Balances        []WalletCurrencyBalance
	WalletRemain    float64
	GiftPoints      float64
	OverdraftPoints float64
}

type Currency struct {
	Code          string
	PointsPerUnit int64
	Enabled       bool
}

type BillingRepository interface {
	CreateRechargeOrder(ctx context.Context, order RechargeOrder) error
	GetRechargeOrder(ctx context.Context, id uuid.UUID) (*RechargeOrder, error)
	ListRechargeOrders(ctx context.Context, companyID uuid.UUID) ([]RechargeOrder, error)
	ConfirmRechargeWithLot(ctx context.Context, order RechargeOrder, lot RechargeLot) error
	ListActiveLotsFIFO(ctx context.Context, companyID uuid.UUID, fifoHeadID *uuid.UUID) ([]RechargeLot, error)
	UpdateLotRemaining(ctx context.Context, lot RechargeLot) error
	GetLotByID(ctx context.Context, lotID uuid.UUID) (*RechargeLot, error)
	ExpandOverdraftLot(ctx context.Context, companyID uuid.UUID, billingCurrency string, pointsDelta float64) (*RechargeLot, error)
	ExpireMockLots(ctx context.Context, companyID uuid.UUID) (int64, error)
	SumActiveLotsRemaining(ctx context.Context, companyID uuid.UUID) (float64, error)
	AggregateWallet(ctx context.Context, companyID uuid.UUID) (WalletAggregate, error)
	GetCurrency(ctx context.Context, code string) (*Currency, error)
}
