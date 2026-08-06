package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type syncVersionsRepo struct {
	db dbQuerier
}

func (r *syncVersionsRepo) Increment(ctx context.Context, companyID uuid.UUID, typ string) (int, error) {
	var v int
	err := r.db.QueryRow(ctx, `
		INSERT INTO sync_versions (company_id, type, version)
		VALUES ($1, $2, 1)
		ON CONFLICT (company_id, type)
		DO UPDATE SET version = sync_versions.version + 1
		RETURNING version
	`, companyID, typ).Scan(&v)
	return v, err
}

func (r *syncVersionsRepo) Set(ctx context.Context, companyID uuid.UUID, typ string, version int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sync_versions (company_id, type, version)
		VALUES ($1, $2, $3)
		ON CONFLICT (company_id, type)
		DO UPDATE SET version = $3
	`, companyID, typ, version)
	return err
}

func (r *syncVersionsRepo) Get(ctx context.Context, companyID uuid.UUID, typ string) (int, error) {
	var v int
	err := r.db.QueryRow(ctx,
		`SELECT version FROM sync_versions WHERE company_id = $1 AND type = $2`,
		companyID, typ,
	).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return v, err
}

func (r *syncVersionsRepo) GetVersions(ctx context.Context, companyID uuid.UUID) (map[string]int, map[string]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT company_id, type, version FROM sync_versions WHERE company_id IN ($1, $2)`,
		store.GlobalSyncVersion, companyID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	global := make(map[string]int)
	company := make(map[string]int)
	for rows.Next() {
		var cid uuid.UUID
		var typ string
		var v int
		if err := rows.Scan(&cid, &typ, &v); err != nil {
			return nil, nil, err
		}
		if cid == store.GlobalSyncVersion {
			global[typ] = v
		} else {
			company[typ] = v
		}
	}
	return global, company, rows.Err()
}

var _ store.SyncVersionRepository = (*syncVersionsRepo)(nil)
