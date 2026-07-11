package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetWallet(ctx context.Context) (WalletView, error)
	ListRechargeRecords(ctx context.Context) ([]RechargeRecord, error)
	PlatformRecharge(ctx context.Context, companyID int64, amount float64, operatorID string) error
	PlatformGift(ctx context.Context, companyID int64, points float64, operatorID string) error
	PlatformAdjust(ctx context.Context, companyID int64, points float64, amountDisplay float64, operatorID string) error
	CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID string) (store.RechargeOrder, error)
	ConfirmPayment(ctx context.Context, orderID string) error
	SyncCompanyWallet(ctx context.Context, companyID int64) error
	ReconcileWalletDrift(ctx context.Context) error
}

type service struct {
	cfg           config.Config
	store         store.Store
	reader        domainusage.Reader
	client        adminport.Port
	wallet        company.WalletService
	rebalanceAxis func(ctx context.Context, companyID int64) error
	enqueueSync   func(ctx context.Context, companyID int64) error
}

func NewService(
	cfg config.Config,
	st store.Store,
	reader domainusage.Reader,
	client adminport.Port,
	wallet company.WalletService,
	rebalanceAxis func(ctx context.Context, companyID int64) error,
	enqueueSync func(ctx context.Context, companyID int64) error,
) Service {
	return &service{
		cfg: cfg, store: st, reader: reader, client: client, wallet: wallet,
		rebalanceAxis: rebalanceAxis, enqueueSync: enqueueSync,
	}
}

func (s *service) PlatformRecharge(ctx context.Context, companyID int64, amount float64, operatorID string) error {
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmPaidRecharge(ctx, amount, store.RechargeSourcePlatform,
		fmt.Sprintf("platform:%s", operatorID), nil); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.recharge", operatorID,
		fmt.Sprintf("company:%d", companyID), fmt.Sprintf("amount=%.2f", amount))
}

func (s *service) PlatformGift(ctx context.Context, companyID int64, points float64, operatorID string) error {
	if points <= 0 {
		return fmt.Errorf("gift points must be positive")
	}
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmGiftLot(ctx, points, fmt.Sprintf("platform-gift:%s", operatorID)); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.gift", operatorID,
		fmt.Sprintf("company:%d", companyID), fmt.Sprintf("points=%.0f", points))
}

func (s *service) PlatformAdjust(ctx context.Context, companyID int64, points float64, amountDisplay float64, operatorID string) error {
	if points <= 0 {
		return fmt.Errorf("adjust points must be positive")
	}
	if amountDisplay < 0 {
		return fmt.Errorf("adjust amount display must be non-negative")
	}
	ctx = company.WithContext(ctx, company.Context{CompanyID: companyID})
	if err := s.confirmAdjustLot(ctx, points, amountDisplay, fmt.Sprintf("platform-adjust:%s", operatorID)); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.adjust", operatorID,
		fmt.Sprintf("company:%d", companyID), fmt.Sprintf("points=%.0f amount=%.2f", points, amountDisplay))
}

func (s *service) CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID string) (store.RechargeOrder, error) {
	companyID := company.CompanyID(ctx)
	now := time.Now().UTC()
	orderID := fmt.Sprintf("rch-%d-%d", companyID, now.UnixNano())
	key := idempotencyKey
	ppu := DefaultPointsPerUnit()
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Currency: "CNY",
		PointsPerUnit: ppu, PointsGranted: PointsGrantedFromAmount(amount, ppu),
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
