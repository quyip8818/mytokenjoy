package billing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
)

type WalletCurrencyView struct {
	Currency      string  `json:"currency"`
	Balance       float64 `json:"balance"`
	TotalTopup    float64 `json:"totalTopup"`
	TotalConsumed float64 `json:"totalConsumed"`
}

type WalletView struct {
	CompanyID       uuid.UUID            `json:"companyId"`
	BillingCurrency string               `json:"billingCurrency"`
	Balances        []WalletCurrencyView `json:"balances"`
	WalletRemain    int64                `json:"walletRemain"`
	GiftQuota       int64                `json:"giftQuota"`
	OverdraftQuota  int64                `json:"overdraftQuota"`
	TotalRequests   int64                `json:"totalRequests"`
}

func PrimaryWalletBalance(view WalletView) float64 {
	for _, b := range view.Balances {
		if b.Currency == view.BillingCurrency {
			return b.Balance
		}
	}
	if len(view.Balances) > 0 {
		return view.Balances[0].Balance
	}
	return 0
}

func (s *service) GetWallet(ctx context.Context) (WalletView, error) {
	companyCtx, ok := company.FromContext(ctx)
	if !ok {
		return WalletView{}, fmt.Errorf("company context required")
	}
	agg, err := s.store.Billing().AggregateWallet(ctx, companyCtx.CompanyID)
	if err != nil {
		return WalletView{}, err
	}
	view := WalletView{
		CompanyID:       companyCtx.CompanyID,
		BillingCurrency: agg.BillingCurrency,
		WalletRemain:    agg.WalletRemain,
		GiftQuota:       agg.GiftQuota,
		OverdraftQuota:  agg.OverdraftQuota,
	}
	for _, b := range agg.Balances {
		view.Balances = append(view.Balances, WalletCurrencyView{
			Currency: b.Currency, Balance: b.Balance,
			TotalTopup: b.TotalTopup, TotalConsumed: b.TotalConsumed,
		})
	}
	requests, err := s.lifetimeRequestCount(ctx)
	if err != nil {
		return WalletView{}, err
	}
	view.TotalRequests = requests
	return view, nil
}
