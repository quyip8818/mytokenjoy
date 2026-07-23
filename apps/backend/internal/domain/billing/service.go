package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
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
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg         config.Config
	store       Store
	reader      domainusage.Reader
	quotaSyncer QuotaSyncer
}

func NewService(
	cfg config.Config,
	st Store,
	reader domainusage.Reader,
	quotaSyncer QuotaSyncer,
) Service {
	return &service{
		cfg: cfg, store: st, reader: reader,
		quotaSyncer: quotaSyncer,
	}
}

// syncQuotaToNewAPI is the PreCreditFunc called before CreditFromLot commits.
// It syncs quota to the external NewAPI gateway so the user won't be rejected.
//
// Returns error to abort the recharge if NewAPI is unreachable — this ensures
// we never have local balance that the user can't actually spend.
func (s *service) syncQuotaToNewAPI(ctx context.Context, lot store.RechargeLot) error {
	if lot.LotKind == store.LotKindOverdraft {
		return nil
	}
	if s.cfg.IsProductionDeploy() && lot.LotKind == store.LotKindMock {
		return nil // prod trial/demo gets a one-time large quota at company creation
	}
	if s.quotaSyncer == nil {
		return nil
	}
	walletUserID, ok := company.ResolveNewAPIWalletCompanyID(ctx, s.store.Company())
	if !ok {
		return nil // company not yet provisioned on NewAPI; skip silently
	}
	return s.quotaSyncer.ManageUser(ctx, walletUserID, "add_quota", lot.QuotaGranted)
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
