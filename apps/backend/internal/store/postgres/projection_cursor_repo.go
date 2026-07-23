package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type projectionCursorRepo struct {
	db dbQuerier
}

func newProjectionCursorRepo(db dbQuerier) store.ProjectionCursorRepository {
	return &projectionCursorRepo{db: db}
}

const projectionCursorsTable = "projection_cursors"

func (r *projectionCursorRepo) Get(ctx context.Context, stream string) (*store.ProjectionCursor, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT company_id, stream, last_occurred_at, last_ledger_id
		FROM %s
		WHERE company_id = $1 AND stream = $2
	`, projectionCursorsTable), companyID, stream)

	var cursor store.ProjectionCursor
	err := row.Scan(&cursor.CompanyID, &cursor.Stream, &cursor.LastOccurredAt, &cursor.LastLedgerID)
	if err == pgx.ErrNoRows {
		return &store.ProjectionCursor{CompanyID: companyID, Stream: stream}, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func (r *projectionCursorRepo) GetForUpdate(ctx context.Context, stream string) (*store.ProjectionCursor, error) {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (company_id, stream)
		VALUES ($1, $2)
		ON CONFLICT (company_id, stream) DO NOTHING
	`, projectionCursorsTable), companyID, stream); err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT company_id, stream, last_occurred_at, last_ledger_id
		FROM %s
		WHERE company_id = $1 AND stream = $2
		FOR UPDATE
	`, projectionCursorsTable), companyID, stream)

	var cursor store.ProjectionCursor
	err := row.Scan(&cursor.CompanyID, &cursor.Stream, &cursor.LastOccurredAt, &cursor.LastLedgerID)
	if err == pgx.ErrNoRows {
		return &store.ProjectionCursor{CompanyID: companyID, Stream: stream}, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func (r *projectionCursorRepo) Advance(ctx context.Context, stream string, lastOccurredAt time.Time, lastLedgerID uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (company_id, stream, last_occurred_at, last_ledger_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (company_id, stream) DO UPDATE SET
			last_occurred_at = EXCLUDED.last_occurred_at,
			last_ledger_id = EXCLUDED.last_ledger_id,
			updated_at = NOW()
	`, projectionCursorsTable), companyID, stream, lastOccurredAt.UTC(), lastLedgerID)
	return err
}
