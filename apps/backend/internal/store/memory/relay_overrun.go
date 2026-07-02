package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryRelayRepo) EnqueueOverrun(ctx context.Context, payload json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	companyID := store.CompanyID(ctx)
	r.data.overrun = append(r.data.overrun, store.OverrunQueueEntry{
		ID:        fmt.Sprintf("ovr-%d-%d", companyID, time.Now().UnixNano()),
		CompanyID: companyID,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
	})
	return nil
}

func (r *memoryRelayRepo) ClaimPendingOverrun(ctx context.Context, limit int) ([]store.OverrunQueueEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]store.OverrunQueueEntry, 0, limit)
	for _, e := range r.data.overrun {
		if len(out) >= limit {
			break
		}
		if e.Status == store.OutboxStatusPending {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkOverrunDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.overrun {
		if r.data.overrun[i].ID == id {
			r.data.overrun[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}
