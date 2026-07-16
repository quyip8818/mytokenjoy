package billing

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func resolveBillingCurrency(co *store.Company) string {
	if co == nil {
		return common.DefaultBillingCurrency
	}
	return common.ResolveBillingCurrency(co.BillingCurrency)
}

// CurrencyStore is the narrow store surface for currency lookups.
type CurrencyStore interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
}

func lookupCurrencyPPU(ctx context.Context, st CurrencyStore, currency string) (int64, error) {
	cur, err := st.Billing().GetCurrency(ctx, currency)
	if err != nil {
		return 0, err
	}
	if cur == nil {
		return 0, domain.BadRequest(fmt.Sprintf("currency %s is not supported", currency))
	}
	if !cur.Enabled {
		return 0, domain.BadRequest(fmt.Sprintf("currency %s is disabled", currency))
	}
	if cur.PointsPerUnit <= 0 {
		return 0, domain.BadRequest(fmt.Sprintf("currency %s has invalid points_per_unit", currency))
	}
	return cur.PointsPerUnit, nil
}

func (s *service) lookupPointsPerUnit(ctx context.Context, currency string) (int64, error) {
	return lookupCurrencyPPU(ctx, s.store, currency)
}

func (s *service) resolveChargeRate(ctx context.Context, companyID int64) (currency string, ppu int64, err error) {
	return ResolveCompanyChargeRate(ctx, s.store, companyID)
}

// ResolveCompanyChargeRate returns the company's billing currency and points_per_unit.
func ResolveCompanyChargeRate(ctx context.Context, st CurrencyStore, companyID int64) (currency string, ppu int64, err error) {
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil {
		return "", 0, err
	}
	if co == nil {
		return "", 0, domain.NotFound("company not found")
	}
	currency = resolveBillingCurrency(co)
	ppu, err = lookupCurrencyPPU(ctx, st, currency)
	if err != nil {
		return "", 0, err
	}
	return currency, ppu, nil
}
