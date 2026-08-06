package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/common"
)

type Service interface {
	GetWallet(ctx context.Context) (WalletView, error)
	ListRechargeRecords(ctx context.Context) ([]RechargeRecord, error)
	PlatformRecharge(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error
	PlatformGift(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error
	PlatformAdjust(ctx context.Context, companyID uuid.UUID, amount float64, paidAmount float64, operatorID uuid.UUID) error
	CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID uuid.UUID) (store.RechargeOrder, error)
	ConfirmPayment(ctx context.Context, orderID uuid.UUID) error
}

// QuotaSyncer is the minimal interface for syncing quota to NewAPI.
type QuotaSyncer interface {
	ManageUser(ctx context.Context, userID int64, action string, value int64) error
}

// Store is the narrow store surface the billing domain needs.
type Store interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
	Models() store.ModelsRepository
	Audit() store.AuditRepository
	SystemSettings() store.SystemSettingsRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	store       Store
	reader      domainusage.Reader
	quotaSyncer QuotaSyncer
}

func NewService(
	st Store,
	reader domainusage.Reader,
	quotaSyncer QuotaSyncer,
) Service {
	return &service{
		store: st, reader: reader,
		quotaSyncer: quotaSyncer,
	}
}

func (s *service) PlatformRecharge(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error {
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmPaidRecharge(ctx, amount, store.RechargeSourcePlatform,
		operatorID, nil); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.recharge", operatorID,
		fmt.Sprintf("company:%s", companyID), fmt.Sprintf("amount=%.2f", amount))
}

func (s *service) PlatformGift(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error {
	if amount <= 0 {
		return fmt.Errorf("gift amount must be positive")
	}
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmGiftLot(ctx, amount, operatorID); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.gift", operatorID,
		fmt.Sprintf("company:%s", companyID), fmt.Sprintf("amount=%.2f", amount))
}

func (s *service) PlatformAdjust(ctx context.Context, companyID uuid.UUID, amount float64, paidAmount float64, operatorID uuid.UUID) error {
	if amount <= 0 {
		return fmt.Errorf("adjust amount must be positive")
	}
	if paidAmount < 0 {
		return fmt.Errorf("adjust paid amount must be non-negative")
	}
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmAdjustLot(ctx, amount, paidAmount, operatorID); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.adjust", operatorID,
		fmt.Sprintf("company:%s", companyID), fmt.Sprintf("amount=%.2f paidAmount=%.2f", amount, paidAmount))
}

func (s *service) CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID uuid.UUID) (store.RechargeOrder, error) {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return store.RechargeOrder{}, err
	}
	now := time.Now().UTC()
	orderID := uuid.Must(uuid.NewV7())
	key := idempotencyKey
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Currency: currency,
		QuotaPerUnit: ppu, QuotaGranted: common.MoneyToQuota(amount, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusPending, CreatedBy: memberID,
		DisplayOrderID: formatDisplayOrderID(now),
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedAt:      now, UpdatedAt: now,
	}
	if err := s.store.Billing().CreateRechargeOrder(ctx, order); err != nil {
		return store.RechargeOrder{}, err
	}
	return order, nil
}
