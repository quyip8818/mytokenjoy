package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/redemption"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/quota"
)

// RedeemResult is the API response payload for a successful redemption.
type RedeemResult struct {
	FaceValue    float64 `json:"faceValue"`
	Currency     string  `json:"currency"`
	QuotaGranted int64   `json:"quotaGranted"`
}

// RedeemCode validates and redeems a code, crediting the current company's wallet.
func (s *service) RedeemCode(ctx context.Context, rawCode string, memberID uuid.UUID) (RedeemResult, error) {
	companyID := company.CompanyID(ctx)
	code := redemption.NormalizeCode(rawCode)

	var result RedeemResult
	var lot store.RechargeLot
	var newRemain int64

	err := s.store.WithTx(ctx, func(tx store.Store) error {
		// 1. Lock the redemption code row.
		rc, err := tx.Redemption().GetCodeForUpdate(ctx, code)
		if err != nil {
			return err
		}
		if rc == nil {
			return domain.ValidationCode("INVALID_REDEMPTION_CODE", "兑换码无效")
		}

		// 2. Validate state.
		if rc.Status == store.RedemptionStatusDisabled {
			return domain.ValidationCode("CODE_DISABLED", "该兑换码已被禁用")
		}
		if rc.Status != store.RedemptionStatusUnused {
			return domain.ValidationCode("CODE_ALREADY_USED", "该兑换码已被使用")
		}
		if rc.ExpiresAt.Before(time.Now().UTC()) {
			return domain.ValidationCode("CODE_EXPIRED", "该兑换码已过期")
		}

		// 3. Resolve charge rate for the company.
		currency, ppu, err := ResolveCompanyChargeRate(ctx, tx, companyID)
		if err != nil {
			return err
		}
		// ponytail: v1 assumes code currency == company billing currency. Cross-currency upgrade path: convert face_value.
		quotaGranted := quota.MoneyToQuota(rc.FaceValue, ppu)

		// 4. Create RechargeOrder + Lot via standard CreditFromLot path.
		now := time.Now().UTC()
		orderID := uuid.Must(uuid.NewV7())
		order := store.RechargeOrder{
			ID: orderID, CompanyID: companyID, Amount: 0, Currency: currency,
			QuotaPerUnit: ppu, QuotaGranted: quotaGranted,
			Source: store.RechargeSourceRedemption, LotKind: store.LotKindGift,
			Status: store.RechargeStatusConfirmed, CreatedBy: memberID,
			CreatedAt: now, UpdatedAt: now,
		}
		lot = BuildLot(order, currency, store.LotKindGift, 0)
		newRemain, err = billinglot.CreditFromLot(ctx, tx, order, lot, lot.QuotaGranted)
		if err != nil {
			return err
		}

		// 5. Mark code as used.
		if err := tx.Redemption().MarkUsed(ctx, rc.ID, companyID, memberID, orderID); err != nil {
			return err
		}

		result = RedeemResult{
			FaceValue:    rc.FaceValue,
			Currency:     currency,
			QuotaGranted: quotaGranted,
		}
		return nil
	})
	if err != nil {
		return RedeemResult{}, err
	}

	// Post-commit best-effort: audit transaction + wallet sync.
	s.recordCreditTransaction(ctx, companyID, lot, "redemption", memberID)
	s.syncWalletBestEffort(ctx, companyID, newRemain)

	return result, nil
}
