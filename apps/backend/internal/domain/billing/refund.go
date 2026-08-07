package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/quota"
)

func (s *service) PlatformRefund(ctx context.Context, companyID uuid.UUID, lotID uuid.UUID, amount float64, operatorID uuid.UUID) error {
	if amount <= 0 {
		return domain.BadRequest("refund amount must be positive")
	}

	var newWalletRemain int64
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		// Lock company row — serializes with ConsumeLots.
		co, err := tx.Company().LockForUpdate(ctx, companyID)
		if err != nil {
			return err
		}
		if co == nil {
			return domain.NotFound("company not found")
		}

		// Read lot inside lock.
		lot, err := tx.Billing().GetLotByID(ctx, lotID)
		if err != nil {
			return err
		}
		if lot == nil {
			return domain.NotFound("lot not found")
		}
		if lot.CompanyID != companyID {
			return domain.Forbidden("lot does not belong to this company")
		}
		if lot.Status != store.LotStatusActive {
			return domain.BadRequest("can only refund active lots")
		}
		if lot.LotKind == store.LotKindOverdraft {
			return domain.BadRequest("cannot refund overdraft lots")
		}

		refundQuota := quota.MoneyToQuota(amount, lot.QuotaPerUnit)
		if refundQuota <= 0 {
			return domain.BadRequest("refund amount too small")
		}
		if refundQuota > lot.QuotaRemaining {
			return domain.BadRequest(fmt.Sprintf("refund quota %d exceeds lot remaining %d", refundQuota, lot.QuotaRemaining))
		}

		// Deduct lot.
		lot.QuotaRemaining -= refundQuota
		if lot.QuotaRemaining <= 0 {
			lot.Status = store.LotStatusExhausted
		}
		if err := tx.Billing().UpdateLotRemaining(ctx, *lot); err != nil {
			return err
		}

		// Recalculate FIFO head if needed.
		var newHead *uuid.UUID
		if co.FIFOHeadLotID != nil && *co.FIFOHeadLotID == lotID && lot.Status == store.LotStatusExhausted {
			// Find next active lot.
			activeLots, err := tx.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
			if err != nil {
				return err
			}
			if len(activeLots) > 0 {
				newHead = &activeLots[0].ID
			}
		}

		// Wallet delta.
		newWalletRemain = co.WalletRemainQuota - refundQuota
		if newWalletRemain < 0 {
			newWalletRemain = 0
		}
		if err := tx.Company().SetWalletRemainQuota(ctx, companyID, newWalletRemain, newHead); err != nil {
			return err
		}

		// Insert lot_transaction.
		now := time.Now().UTC()
		txRecord := store.LotTransaction{
			ID:              uuid.Must(uuid.NewV7()),
			CompanyID:       companyID,
			LotID:           lotID,
			Action:          "refund",
			QuotaDelta:      -refundQuota,
			QuotaPerUnit:    lot.QuotaPerUnit,
			MoneyAmount:     amount,
			BillingCurrency: lot.BillingCurrency,
			RemainingAfter:  lot.QuotaRemaining,
			Source:          "platform",
			LotKind:         lot.LotKind,
			OperatorID:      &operatorID,
			Note:            fmt.Sprintf("退费 ¥%.2f", amount),
			CreatedAt:       now,
		}
		if err := tx.Billing().InsertLotTransaction(ctx, txRecord); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Post-commit: bump version + sync.
	s.syncWalletBestEffort(ctx, companyID, newWalletRemain)

	// Audit log.
	_ = company.AppendPlatformOperationLog(ctx, s.store, companyID, "platform.company.refund", operatorID,
		fmt.Sprintf("company:%s lot:%s", companyID, lotID), fmt.Sprintf("amount=%.2f", amount))

	return nil
}
