package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetWallet(ctx context.Context) (WalletView, error)
	PlatformRecharge(ctx context.Context, companyID int64, amount float64, operatorID string) error
	CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID string) (store.RechargeOrder, error)
	ConfirmPayment(ctx context.Context, orderID string) error
}

type WalletView struct {
	Balance     float64 `json:"balance"`
	Allocatable float64 `json:"allocatable"`
	Currency    string  `json:"currency"`
	CompanyID   int64   `json:"companyId"`
}

type service struct {
	cfg           config.Config
	store         store.Store
	client        newapi.AdminClient
	wallet        company.WalletService
	rebalanceAxis func(ctx context.Context, companyID int64) error
}

func NewService(
	cfg config.Config,
	st store.Store,
	client newapi.AdminClient,
	wallet company.WalletService,
	rebalanceAxis func(ctx context.Context, companyID int64) error,
) Service {
	return &service{cfg: cfg, store: st, client: client, wallet: wallet, rebalanceAxis: rebalanceAxis}
}

func (s *service) GetWallet(ctx context.Context) (WalletView, error) {
	companyCtx, ok := company.FromContext(ctx)
	if !ok {
		return WalletView{}, fmt.Errorf("company context required")
	}
	view := WalletView{Currency: "CNY", CompanyID: companyCtx.CompanyID}
	if companyCtx.NewAPIWalletUserID <= 0 {
		return view, nil
	}
	quota, err := s.wallet.AvailableQuota(ctx, companyCtx.NewAPIWalletUserID)
	if err != nil {
		return WalletView{}, err
	}
	balance := newapi.FromNewAPIUnits(quota, nil, nil)
	view.Balance = balance
	view.Allocatable = balance
	return view, nil
}

func (s *service) PlatformRecharge(ctx context.Context, companyID int64, amount float64, operatorID string) error {
	if err := s.executeRecharge(ctx, companyID, amount, store.RechargeSourcePlatform,
		fmt.Sprintf("platform:%s", operatorID), nil); err != nil {
		return err
	}
	return company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.recharge", operatorID,
		fmt.Sprintf("company:%d", companyID), fmt.Sprintf("amount=%.2f", amount))
}

func (s *service) CreateSelfRecharge(ctx context.Context, amount float64, idempotencyKey string, memberID string) (store.RechargeOrder, error) {
	companyID := company.CompanyID(ctx)
	orderID := fmt.Sprintf("rch-%d-%d", companyID, time.Now().UnixNano())
	now := time.Now().UTC()
	key := idempotencyKey
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Source: store.RechargeSourceSelf,
		IdempotencyKey: &key, Status: store.RechargeStatusPending, CreatedBy: memberID,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.Billing().CreateRechargeOrder(ctx, order); err != nil {
		return store.RechargeOrder{}, err
	}
	return order, nil
}

func (s *service) ConfirmPayment(ctx context.Context, orderID string) error {
	order, err := s.store.Billing().GetRechargeOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status == store.RechargeStatusToppedUp {
		return nil
	}
	if err := s.store.Billing().UpdateRechargeStatus(ctx, orderID, store.RechargeStatusPaid, nil); err != nil {
		return err
	}
	return s.topUpAndFinish(ctx, *order)
}

func (s *service) executeRecharge(ctx context.Context, companyID int64, amount float64, source, createdBy string, idempotencyKey *string) error {
	orderID := fmt.Sprintf("rch-%d-%d", companyID, time.Now().UnixNano())
	now := time.Now().UTC()
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Source: source,
		IdempotencyKey: idempotencyKey, Status: store.RechargeStatusPaid,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.Billing().CreateRechargeOrder(ctx, order); err != nil {
		return err
	}
	return s.topUpAndFinish(ctx, order)
}

func (s *service) topUpAndFinish(ctx context.Context, order store.RechargeOrder) error {
	co, err := s.store.Company().GetByID(ctx, order.CompanyID)
	if err != nil || co == nil || co.NewAPIWalletUserID == nil {
		_ = s.store.Billing().UpdateRechargeStatus(ctx, order.ID, store.RechargeStatusFailed, nil)
		return fmt.Errorf("company wallet not configured")
	}
	units := newapi.ToNewAPIUnits(order.Amount, nil, nil)
	if s.cfg.NewAPIEnabled && s.client != nil {
		if err := s.client.TopUp(ctx, newapi.TopUpRequest{
			UserID: *co.NewAPIWalletUserID,
			Quota:  units,
			Remark: fmt.Sprintf("recharge %s", order.ID),
		}); err != nil {
			_ = s.store.Billing().UpdateRechargeStatus(ctx, order.ID, store.RechargeStatusFailed, nil)
			return err
		}
	}
	ref := order.ID
	if err := s.store.Billing().UpdateRechargeStatus(ctx, order.ID, store.RechargeStatusToppedUp, &ref); err != nil {
		return err
	}
	if s.rebalanceAxis != nil {
		companyCtx := company.WithContext(ctx, company.Context{
			CompanyID:          co.ID,
			NewAPIWalletUserID: *co.NewAPIWalletUserID,
			Status:             co.Status,
		})
		_ = s.rebalanceAxis(companyCtx, co.ID)
	}
	return nil
}
