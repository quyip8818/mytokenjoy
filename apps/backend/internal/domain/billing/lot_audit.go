package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/quota"
)

type LotTransactionView struct {
	ID             uuid.UUID  `json:"id"`
	Action         string     `json:"action"`
	QuotaDelta     int64      `json:"quotaDelta"`
	MoneyAmount    float64    `json:"moneyAmount"`
	RemainingAfter int64      `json:"remainingAfter"`
	Source         string     `json:"source"`
	OperatorID     *uuid.UUID `json:"operatorId,omitempty"`
	Note           string     `json:"note"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type LotAuditEntry struct {
	ID              uuid.UUID            `json:"id"`
	OrderID         uuid.UUID            `json:"orderId"`
	LotKind         string               `json:"lotKind"`
	BillingCurrency string               `json:"billingCurrency"`
	QuotaPerUnit    int64                `json:"quotaPerUnit"`
	QuotaGranted    int64                `json:"quotaGranted"`
	QuotaRemaining  int64                `json:"quotaRemaining"`
	PaidAmount      float64              `json:"paidAmount"`
	Status          string               `json:"status"`
	CreatedAt       time.Time            `json:"createdAt"`
	Transactions    []LotTransactionView `json:"transactions"`
}

// ListLots returns all lots + transactions for the current company (billing:read).
func (s *service) ListLots(ctx context.Context) ([]LotAuditEntry, error) {
	companyID := company.CompanyID(ctx)
	return s.buildLotAudit(ctx, companyID)
}

// PlatformListLots returns all lots + transactions for a given company (platform admin).
func (s *service) PlatformListLots(ctx context.Context, companyID uuid.UUID) ([]LotAuditEntry, error) {
	return s.buildLotAudit(ctx, companyID)
}

func (s *service) buildLotAudit(ctx context.Context, companyID uuid.UUID) ([]LotAuditEntry, error) {
	lots, err := s.store.Billing().ListAllLots(ctx, companyID)
	if err != nil {
		return nil, err
	}
	txs, err := s.store.Billing().ListLotTransactions(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Group transactions by lot_id.
	txByLot := make(map[uuid.UUID][]LotTransactionView, len(txs))
	for _, t := range txs {
		txByLot[t.LotID] = append(txByLot[t.LotID], LotTransactionView{
			ID:             t.ID,
			Action:         t.Action,
			QuotaDelta:     t.QuotaDelta,
			MoneyAmount:    t.MoneyAmount,
			RemainingAfter: t.RemainingAfter,
			Source:         t.Source,
			OperatorID:     t.OperatorID,
			Note:           t.Note,
			CreatedAt:      t.CreatedAt,
		})
	}

	entries := make([]LotAuditEntry, 0, len(lots))
	for _, lot := range lots {
		entry := LotAuditEntry{
			ID:              lot.ID,
			OrderID:         lot.RechargeOrderID,
			LotKind:         lot.LotKind,
			BillingCurrency: lot.BillingCurrency,
			QuotaPerUnit:    lot.QuotaPerUnit,
			QuotaGranted:    lot.QuotaGranted,
			QuotaRemaining:  lot.QuotaRemaining,
			PaidAmount:      lot.PaidAmount,
			Status:          lot.Status,
			CreatedAt:       lot.CreatedAt,
			Transactions:    txByLot[lot.ID],
		}
		if entry.Transactions == nil {
			entry.Transactions = []LotTransactionView{}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// RecordCreditTransaction inserts a credit lot_transaction after a successful CreditFromLot.
// ponytail: helper called from confirm paths. Not on Service interface — internal only.
func (s *service) recordCreditTransaction(ctx context.Context, companyID uuid.UUID, lot store.RechargeLot, source string, operatorID uuid.UUID) {
	now := time.Now().UTC()
	moneyAmount := quota.QuotaToMoney(lot.QuotaGranted, lot.QuotaPerUnit)
	tx := store.LotTransaction{
		ID:              uuid.Must(uuid.NewV7()),
		CompanyID:       companyID,
		LotID:           lot.ID,
		Action:          "credit",
		QuotaDelta:      lot.QuotaGranted,
		QuotaPerUnit:    lot.QuotaPerUnit,
		MoneyAmount:     moneyAmount,
		BillingCurrency: lot.BillingCurrency,
		RemainingAfter:  lot.QuotaRemaining,
		Source:          source,
		LotKind:         lot.LotKind,
		OperatorID:      &operatorID,
		Note:            "",
		CreatedAt:       now,
	}
	// Best-effort: don't fail the recharge if transaction logging fails.
	_ = s.store.Billing().InsertLotTransaction(ctx, tx)
}
