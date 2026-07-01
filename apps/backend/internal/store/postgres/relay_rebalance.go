package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *relayRepo) EnqueueRebalance(ctx context.Context, axisKind, axisID string) error {
	companyID := store.CompanyID(ctx)
	id := fmt.Sprintf("rb-%d-%s-%s-%d", companyID, axisKind, axisID, time.Now().UnixNano())
	_, err := r.db.Exec(ctx, `
		INSERT INTO rebalance_queue (id, company_id, axis_kind, axis_id, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (company_id, axis_kind, axis_id, status) DO NOTHING
	`, id, companyID, axisKind, axisID, store.OutboxStatusPending)
	return err
}

func (r *relayRepo) ClaimPendingRebalance(ctx context.Context, limit int) ([]store.RebalanceQueueEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, axis_kind, axis_id, status
		FROM rebalance_queue
		WHERE status = $1
		ORDER BY created_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, store.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.RebalanceQueueEntry, 0)
	for rows.Next() {
		var e store.RebalanceQueueEntry
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.AxisKind, &e.AxisID, &e.Status); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkRebalanceDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE rebalance_queue SET status = $2 WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}
