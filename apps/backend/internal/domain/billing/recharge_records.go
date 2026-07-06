// TODO(real): persist payment method, invoice status, and display order numbers via billing domain.
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

type rechargeOverlay struct {
	OrderID       string
	Method        string
	InvoiceStatus string
}

func seedRechargeOverlays() map[string]rechargeOverlay {
	return map[string]rechargeOverlay{
		"tu-1": {OrderID: "ORD202606190001", Method: "alipay", InvoiceStatus: "none"},
		"tu-2": {OrderID: "ORD202606180002", Method: "wechat", InvoiceStatus: "applied"},
		"tu-3": {OrderID: "ORD202606150003", Method: "alipay", InvoiceStatus: "issued"},
		"tu-4": {OrderID: "ORD202606120004", Method: "wechat", InvoiceStatus: "none"},
		"tu-5": {OrderID: "ORD202606100005", Method: "alipay", InvoiceStatus: "issued"},
	}
}

func (s *service) ListRechargeRecords(ctx context.Context) ([]RechargeRecord, error) {
	companyID := company.CompanyID(ctx)
	orders, err := s.store.Billing().ListRechargeOrders(ctx, companyID)
	if err != nil {
		return nil, err
	}
	overlays := seedRechargeOverlays()
	records := make([]RechargeRecord, 0, len(orders))
	for _, order := range orders {
		overlay := overlays[order.ID]
		records = append(records, mapRechargeOrder(order, overlay))
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func mapRechargeOrder(order store.RechargeOrder, overlay rechargeOverlay) RechargeRecord {
	method := overlay.Method
	if method == "" {
		method = "alipay"
	}
	invoiceStatus := overlay.InvoiceStatus
	if invoiceStatus == "" {
		invoiceStatus = "none"
	}
	orderID := overlay.OrderID
	if orderID == "" {
		orderID = order.ID
	}
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
		OrderID:       orderID,
		Method:        method,
		Amount:        order.Amount,
		PaidAmount:    paidAmount,
		InvoiceStatus: invoiceStatus,
		Status:        apiStatus,
		CreatedAt:     order.CreatedAt.In(loc).Format("2006-01-02 15:04:05"),
	}
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
