//go:build testhook

package authz_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/identity/authz"
	"github.com/tokenjoy/backend/internal/store"
)

func testChargeRate(st store.Store) authz.ChargeRateResolver {
	return func(ctx context.Context, companyID uuid.UUID) (string, int64, error) {
		return billing.ResolveCompanyChargeRate(ctx, st, companyID)
	}
}
