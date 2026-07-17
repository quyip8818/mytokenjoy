package billing

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

type RechargeRecord struct {
	ID            uuid.UUID `json:"id"`
	OrderID       string    `json:"orderId"`
	Method        string    `json:"method"`
	Amount        float64   `json:"amount"`
	PaidAmount    float64   `json:"paidAmount"`
	InvoiceStatus string    `json:"invoiceStatus"`
	Status        string    `json:"status"`
	CreatedAt     string    `json:"createdAt"`
}

func (s *service) ListRechargeRecords(ctx context.Context) ([]RechargeRecord, error) {
	companyID := company.CompanyID(ctx)
	orders, err := s.store.Billing().ListRechargeOrders(ctx, companyID)
	if err != nil {
		return nil, err
	}
	records := make([]RechargeRecord, 0, len(orders))
	for _, order := range orders {
		if order.Source != store.RechargeSourceSelf {
			continue
		}
		records = append(records, mapRechargeOrder(order))
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func mapRechargeOrder(order store.RechargeOrder) RechargeRecord {
	paidAmount := 0.0
	if order.Status == store.RechargeStatusConfirmed {
		paidAmount = order.Amount
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	return RechargeRecord{
		ID:            order.ID,
		OrderID:       order.DisplayOrderID,
		Method:        order.PaymentMethod,
		Amount:        order.Amount,
		PaidAmount:    paidAmount,
		InvoiceStatus: order.InvoiceStatus,
		Status:        order.Status,
		CreatedAt:     order.CreatedAt.In(loc).Format("2006-01-02 15:04:05"),
	}
}

func formatDisplayOrderID(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	return "ORD" + t.In(loc).Format("20060102150405")
}
