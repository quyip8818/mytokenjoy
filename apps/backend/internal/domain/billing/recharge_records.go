package billing

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type RechargeRecord struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"orderId"`
	Method        string  `json:"method"`
	Amount        float64 `json:"amount"`
	PaidAmount    float64 `json:"paidAmount"`
	InvoiceStatus string  `json:"invoiceStatus"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"createdAt"`
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
	apiStatus := "pending"
	switch order.Status {
	case store.RechargeStatusToppedUp:
		paidAmount = order.Amount
		apiStatus = "success"
	case store.RechargeStatusPaid:
		paidAmount = order.Amount
		apiStatus = "success"
	case store.RechargeStatusPending:
		apiStatus = "pending"
	case store.RechargeStatusFailed:
		apiStatus = "failed"
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
		Status:        apiStatus,
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

func (s *service) walletUsageStats(ctx context.Context) (float64, int64, error) {
	totals, err := s.store.Usage().QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("query usage summary: %w", err)
	}
	return totals.CostCNY, int64(totals.CallCount), nil
}
