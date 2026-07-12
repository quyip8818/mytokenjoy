package store

import (
	"context"
	"time"
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
	ID             string
	CompanyID      int64
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
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RechargeLot struct {
	ID               string
	CompanyID        int64
	RechargeOrderID  string
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

type BillingRepository interface {
	CreateRechargeOrder(ctx context.Context, order RechargeOrder) error
	GetRechargeOrder(ctx context.Context, id string) (*RechargeOrder, error)
	ListRechargeOrders(ctx context.Context, companyID int64) ([]RechargeOrder, error)
	ConfirmRechargeWithLot(ctx context.Context, order RechargeOrder, lot RechargeLot) error
	ListActiveLotsFIFO(ctx context.Context, companyID int64, fifoHeadID *string) ([]RechargeLot, error)
	UpdateLotRemaining(ctx context.Context, lot RechargeLot) error
	GetLotByID(ctx context.Context, lotID string) (*RechargeLot, error)
	ExpandOverdraftLot(ctx context.Context, companyID int64, billingCurrency string, pointsDelta float64) (*RechargeLot, error)
	AggregateWallet(ctx context.Context, companyID int64) (WalletAggregate, error)
}
