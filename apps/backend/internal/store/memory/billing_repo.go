package memory

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryBillingRepo struct {
	store *Store
}

func (r *memoryBillingRepo) CreateRechargeOrder(ctx context.Context, order store.RechargeOrder) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.rechargeOrders == nil {
		r.store.rechargeOrders = make(map[string]store.RechargeOrder)
	}
	if order.IdempotencyKey != nil && *order.IdempotencyKey != "" {
		for _, existing := range r.store.rechargeOrders {
			if existing.CompanyID == order.CompanyID &&
				existing.IdempotencyKey != nil &&
				*existing.IdempotencyKey == *order.IdempotencyKey {
				return fmt.Errorf("duplicate idempotency key")
			}
		}
	}
	r.store.rechargeOrders[order.ID] = order
	return nil
}

func (r *memoryBillingRepo) GetRechargeOrder(ctx context.Context, id string) (*store.RechargeOrder, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	o, ok := r.store.rechargeOrders[id]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	copy := o
	return &copy, nil
}

func (r *memoryBillingRepo) UpdateRechargeStatus(ctx context.Context, id, status string, topupRef *string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	o, ok := r.store.rechargeOrders[id]
	if !ok {
		return fmt.Errorf("order not found")
	}
	o.Status = status
	if topupRef != nil {
		o.NewAPITopupRef = topupRef
	}
	r.store.rechargeOrders[id] = o
	return nil
}

func (r *memoryBillingRepo) ListRechargeOrders(ctx context.Context, companyID int64) ([]store.RechargeOrder, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var orders []store.RechargeOrder
	for _, o := range r.store.rechargeOrders {
		if o.CompanyID == companyID {
			orders = append(orders, o)
		}
	}
	return orders, nil
}

var _ store.BillingRepository = (*memoryBillingRepo)(nil)
