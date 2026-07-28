package platform

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/store"
)

type companyOverviewItem struct {
	ID              uuid.UUID           `json:"id"`
	Name            string              `json:"name"`
	Type            string              `json:"type"`
	Status          string              `json:"status"`
	BillingCurrency string              `json:"billingCurrency"`
	Wallet          companyWalletSummary `json:"wallet"`
	MonthlySpend    float64             `json:"monthlySpend"`
	MemberCount     int                 `json:"memberCount"`
	CreatedAt       time.Time           `json:"createdAt"`
}

type companyWalletSummary struct {
	Balance       float64 `json:"balance"`
	GiftBalance   float64 `json:"giftBalance"`
	Overdraft     float64 `json:"overdraft"`
	TotalTopup    float64 `json:"totalTopup"`
	TotalConsumed float64 `json:"totalConsumed"`
}

func (h *Handler) CompaniesOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	companies, err := h.p.CompanySvc.ListCompanies(ctx)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	// Resolve quotaPerUnit once (all companies use CNY).
	cur, err := h.p.Billing.GetCurrency(ctx, "CNY")
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	var quotaPerUnit int64 = 1
	if cur != nil && cur.QuotaPerUnit > 0 {
		quotaPerUnit = cur.QuotaPerUnit
	}

	// Batch queries.
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	monthlyCosts, err := h.p.PlatformQuery.SumMonthlyCost(ctx, monthStart, monthEnd)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	memberCounts, err := h.p.PlatformQuery.CountActiveMembers(ctx)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	// ponytail: AggregateWallet per-company loop; acceptable for <100 companies.
	// Upgrade path: batch SQL joining company_recharge_lots if >200 companies.
	result := make([]companyOverviewItem, 0, len(companies))
	for _, co := range companies {
		wallet := walletSummaryFor(ctx, h.p.Billing, co, quotaPerUnit)
		result = append(result, companyOverviewItem{
			ID:              co.ID,
			Name:            co.Name,
			Type:            co.Type,
			Status:          co.Status,
			BillingCurrency: co.BillingCurrency,
			Wallet:          wallet,
			MonthlySpend:    monthlyCosts[co.ID],
			MemberCount:     memberCounts[co.ID],
			CreatedAt:       co.CreatedAt,
		})
	}

	httputil.WriteOK(w, result)
}

func walletSummaryFor(ctx context.Context, billing store.BillingRepository, co store.Company, quotaPerUnit int64) companyWalletSummary {
	agg, err := billing.AggregateWallet(ctx, co.ID)
	if err != nil {
		return companyWalletSummary{}
	}

	var balance, totalTopup, totalConsumed float64
	var found bool
	for _, b := range agg.Balances {
		if b.Currency == co.BillingCurrency {
			balance = b.Balance
			totalTopup = b.TotalTopup
			totalConsumed = b.TotalConsumed
			found = true
			break
		}
	}
	if !found && len(agg.Balances) > 0 {
		balance = agg.Balances[0].Balance
		totalTopup = agg.Balances[0].TotalTopup
		totalConsumed = agg.Balances[0].TotalConsumed
	}

	qpu := float64(quotaPerUnit)
	return companyWalletSummary{
		Balance:       balance,
		GiftBalance:   float64(agg.GiftQuota) / qpu,
		Overdraft:     float64(agg.OverdraftQuota) / qpu,
		TotalTopup:    totalTopup,
		TotalConsumed: totalConsumed,
	}
}
