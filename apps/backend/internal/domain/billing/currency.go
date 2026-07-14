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

func (s *service) lookupPointsPerUnit(ctx context.Context, currency string) (int64, error) {
	cur, err := s.store.Billing().GetCurrency(ctx, currency)
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

func (s *service) resolveChargeRate(ctx context.Context, companyID int64) (currency string, ppu int64, err error) {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return "", 0, err
	}
	if co == nil {
		return "", 0, domain.NotFound("company not found")
	}
	currency = resolveBillingCurrency(co)
	ppu, err = s.lookupPointsPerUnit(ctx, currency)
	if err != nil {
		return "", 0, err
	}
	return currency, ppu, nil
}
