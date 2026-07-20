package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetWallet(ctx context.Context) (WalletView, error)
	ListRechargeRecords(ctx context.Context) ([]RechargeRecord, error)
	PlatformRecharge(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error
	PlatformGift(ctx context.Context, companyID uuid.UUID, amount float64, operatorID uuid.UUID) error
	PlatformAdjust(ctx context.Context, companyID uuid.UUID, amount float64, amountDisplay float64, operatorID uuid.UUID) error
	CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID uuid.UUID) (store.RechargeOrder, error)
	ConfirmPayment(ctx context.Context, orderID uuid.UUID) error
}

// Store is the narrow store surface the billing domain needs.
type Store interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
	Models() store.ModelsRepository
	Audit() store.AuditRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg         config.Config
	store       Store
	reader      domainusage.Reader
	adminClient adminport.Port // nil-safe: no-op when NewAPI is not configured
}

func NewService(
	cfg config.Config,
	st Store,
	reader domainusage.Reader,
	adminClient adminport.Port,
) Service {
	return &service{
		cfg: cfg, store: st, reader: reader, adminClient: adminClient,
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

func (s *service) PlatformAdjust(ctx context.Context, companyID uuid.UUID, amount float64, amountDisplay float64, operatorID uuid.UUID) error {
	if amount <= 0 {
		return fmt.Errorf("adjust amount must be positive")
	}
	if amountDisplay < 0 {
		return fmt.Errorf("adjust amount display must be non-negative")
	}
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmAdjustLot(ctx, amount, amountDisplay, operatorID); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.adjust", operatorID,
		fmt.Sprintf("company:%s", companyID), fmt.Sprintf("amount=%.2f display=%.2f", amount, amountDisplay))
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
		QuotaPerUnit: ppu, QuotaGranted: common.QuotaFromAmount(amount, ppu),
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
