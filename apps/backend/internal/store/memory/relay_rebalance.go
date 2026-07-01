package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryRelayRepo) EnqueueRebalance(ctx context.Context, axisKind, axisID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	companyID := store.CompanyID(ctx)
	for _, e := range r.data.rebalance {
		if e.CompanyID == companyID && e.AxisKind == axisKind && e.AxisID == axisID && e.Status == store.OutboxStatusPending {
			return nil
		}
	}
	r.data.rebalance = append(r.data.rebalance, store.RebalanceQueueEntry{
		ID:        fmt.Sprintf("rb-%d-%s-%s-%d", companyID, axisKind, axisID, time.Now().UnixNano()),
		CompanyID: companyID,
		AxisKind:  axisKind,
		AxisID:    axisID,
		Status:    store.OutboxStatusPending,
	})
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRebalance(ctx context.Context, limit int) ([]store.RebalanceQueueEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]store.RebalanceQueueEntry, 0, limit)
	for _, e := range r.data.rebalance {
		if len(out) >= limit {
			break
		}
		if e.Status == store.OutboxStatusPending {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRebalanceDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.rebalance {
		if r.data.rebalance[i].ID == id {
			r.data.rebalance[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}
