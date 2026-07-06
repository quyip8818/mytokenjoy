package store

import (
	"context"
	"time"
)

const (
	InviteRoleSuperAdmin = "super_admin"

	RechargeStatusPending  = "pending"
	RechargeStatusPaid     = "paid"
	RechargeStatusToppedUp = "topped_up"
	RechargeStatusFailed   = "failed"

	RechargeSourcePlatform = "platform"
	RechargeSourceSelf     = "self"

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
	Source         string
	IdempotencyKey *string
	NewAPITopupRef *string
	Status         string
	DisplayOrderID string
	PaymentMethod  string
	InvoiceStatus  string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	UpdateRechargeStatus(ctx context.Context, id, status string, topupRef *string) error
	ListRechargeOrders(ctx context.Context, companyID int64) ([]RechargeOrder, error)
}
