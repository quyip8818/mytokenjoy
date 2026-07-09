package store

import (
	"context"
	"time"
)

const (
	InviteRoleSuperAdmin = "super_admin"

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
)

type CompanyInvite struct {
	ID         string
	CompanyID  int64
	Email      string
	Role       string
	Token      string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, invite CompanyInvite) error
	GetInviteByToken(ctx context.Context, token string) (*CompanyInvite, error)
	MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error
}

type PlatformOperator struct {
	ID           string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlatformRepository interface {
	GetOperatorByEmail(ctx context.Context, email string) (*PlatformOperator, error)
	GetOperatorByID(ctx context.Context, id string) (*PlatformOperator, error)
	CreateOperator(ctx context.Context, op PlatformOperator) error
	CountOperators(ctx context.Context) (int, error)
}

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
	NewAPISyncRef  *string
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
	BalancePoint    float64
	GiftPoints      float64
	OverdraftPoints float64
}

const (
	InvoiceStatusNone    = "none"
	InvoiceStatusApplied = "applied"
	InvoiceStatusIssued  = "issued"

	PaymentMethodAlipay = "alipay"
	PaymentMethodWechat = "wechat"
)

type BillingRepository interface {
	CreateRechargeOrder(ctx context.Context, order RechargeOrder) error
	GetRechargeOrder(ctx context.Context, id string) (*RechargeOrder, error)
	ListRechargeOrders(ctx context.Context, companyID int64) ([]RechargeOrder, error)
	ConfirmRechargeWithLot(ctx context.Context, order RechargeOrder, lot RechargeLot, balanceDeltaPoint float64) error
	ListActiveLotsFIFO(ctx context.Context, companyID int64, fifoHeadID *string) ([]RechargeLot, error)
	UpdateLotRemaining(ctx context.Context, lot RechargeLot) error
	GetLotByID(ctx context.Context, lotID string) (*RechargeLot, error)
	ExpandOverdraftLot(ctx context.Context, companyID int64, billingCurrency string, pointsDelta float64) (*RechargeLot, error)
	AggregateWallet(ctx context.Context, companyID int64) (WalletAggregate, error)
}
