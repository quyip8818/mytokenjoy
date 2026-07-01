package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryInviteRepo struct {
	store *Store
}

func (r *memoryInviteRepo) CreateInvite(ctx context.Context, invite store.CompanyInvite) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.invites == nil {
		r.store.invites = make(map[string]store.CompanyInvite)
	}
	r.store.invites[invite.Token] = invite
	return nil
}

func (r *memoryInviteRepo) GetInviteByToken(ctx context.Context, token string) (*store.CompanyInvite, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	inv, ok := r.store.invites[token]
	if !ok {
		return nil, fmt.Errorf("invite not found")
	}
	copy := inv
	return &copy, nil
}

func (r *memoryInviteRepo) MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for token, inv := range r.store.invites {
		if inv.ID == id {
			inv.AcceptedAt = &acceptedAt
			r.store.invites[token] = inv
			return nil
		}
	}
	return fmt.Errorf("invite not found")
}

type memoryPlatformRepo struct {
	store *Store
}

func (r *memoryPlatformRepo) GetOperatorByEmail(ctx context.Context, email string) (*store.PlatformOperator, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, op := range r.store.platformOperators {
		if op.Email == email {
			copy := op
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("operator not found")
}

func (r *memoryPlatformRepo) GetOperatorByID(ctx context.Context, id string) (*store.PlatformOperator, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	op, ok := r.store.platformOperators[id]
	if !ok {
		return nil, fmt.Errorf("operator not found")
	}
	copy := op
	return &copy, nil
}

func (r *memoryPlatformRepo) CreateOperator(ctx context.Context, op store.PlatformOperator) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.platformOperators == nil {
		r.store.platformOperators = make(map[string]store.PlatformOperator)
	}
	r.store.platformOperators[op.ID] = op
	return nil
}

func (r *memoryPlatformRepo) CountOperators(ctx context.Context) (int, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return len(r.store.platformOperators), nil
}

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

var (
	_ store.InviteRepository   = (*memoryInviteRepo)(nil)
	_ store.PlatformRepository = (*memoryPlatformRepo)(nil)
	_ store.BillingRepository  = (*memoryBillingRepo)(nil)
)
