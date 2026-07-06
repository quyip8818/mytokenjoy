package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func ApplyRechargeOrders(ctx context.Context, st store.Store) error {
	if _, ok := company.FromContext(ctx); !ok {
		ctx = company.DefaultContext(DefaultCompanyID)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, DefaultCompanyID)
	if err != nil {
		return fmt.Errorf("list recharge orders: %w", err)
	}
	if len(orders) > 0 {
		return nil
	}
	for _, order := range buildSeedRechargeOrders() {
		if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
			return fmt.Errorf("seed recharge order %s: %w", order.ID, err)
		}
	}
	return nil
}

func buildSeedRechargeOrders() []store.RechargeOrder {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	specs := []struct {
		id        string
		amount    float64
		status    string
		createdAt string
	}{
		{"tu-1", 100, store.RechargeStatusToppedUp, "2026-06-19 14:30:00"},
		{"tu-2", 50, store.RechargeStatusToppedUp, "2026-06-18 10:15:00"},
		{"tu-3", 200, store.RechargeStatusToppedUp, "2026-06-15 09:00:00"},
		{"tu-4", 20, store.RechargeStatusPending, "2026-06-12 16:45:00"},
		{"tu-5", 500, store.RechargeStatusToppedUp, "2026-06-10 08:20:00"},
	}
	orders := make([]store.RechargeOrder, 0, len(specs))
	for _, spec := range specs {
		createdAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", spec.createdAt, loc)
		if parseErr != nil {
			createdAt = time.Now().UTC()
		}
		order := store.RechargeOrder{
			ID:        spec.id,
			CompanyID: DefaultCompanyID,
			Amount:    spec.amount,
			Source:    store.RechargeSourceSelf,
			Status:    spec.status,
			CreatedBy: IDMemberAdmin,
			CreatedAt: createdAt.UTC(),
			UpdatedAt: createdAt.UTC(),
		}
		if spec.status == store.RechargeStatusToppedUp {
			ref := spec.id
			order.NewAPITopupRef = &ref
		}
		orders = append(orders, order)
	}
	return orders
}
