package memory

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryRelayRepo) EnqueueRelayOutbox(ctx context.Context, entry store.RelayOutboxEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.relayOutbox = append(r.data.relayOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]store.RelayOutboxEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]store.RelayOutboxEntry, 0, limit)
	for i := range r.data.relayOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.relayOutbox[i]
		if e.Status == store.OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRelayOutboxDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkRelayOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Attempts++
			r.data.relayOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.relayOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) relayOutboxEntry(id string) (store.RelayOutboxEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.data.relayOutbox {
		if entry.ID == id {
			return entry, true
		}
	}
	return store.RelayOutboxEntry{}, false
}

func (r *memoryRelayRepo) GetLastLogID(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data.lastLogID, nil
}

func (r *memoryRelayRepo) SetLastLogID(ctx context.Context, logID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.lastLogID = logID
	return nil
}
