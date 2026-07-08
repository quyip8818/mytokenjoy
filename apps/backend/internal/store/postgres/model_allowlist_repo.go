package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type pgModelAllowlistRepo struct {
	db dbQuerier
}

func (r *pgModelAllowlistRepo) List(ctx context.Context, ownerType, ownerID string) ([]int64, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT model_id
		FROM model_allowlist
		WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
		ORDER BY model_id
	`, companyID, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	modelIDs := make([]int64, 0)
	for rows.Next() {
		var modelID int64
		if err := rows.Scan(&modelID); err != nil {
			return nil, err
		}
		modelIDs = append(modelIDs, modelID)
	}
	return modelIDs, rows.Err()
}

func (r *pgModelAllowlistRepo) Replace(ctx context.Context, ownerType, ownerID string, modelIDs []int64) error {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, `
		DELETE FROM model_allowlist WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
	`, companyID, ownerType, ownerID); err != nil {
		return fmt.Errorf("clear allowlist: %w", err)
	}
	for _, modelID := range modelIDs {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO model_allowlist (company_id, owner_type, owner_id, model_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`, companyID, ownerType, ownerID, modelID); err != nil {
			return fmt.Errorf("insert allowlist row: %w", err)
		}
	}
	return nil
}

func (r *pgModelAllowlistRepo) DeleteByOwner(ctx context.Context, ownerType, ownerID string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		DELETE FROM model_allowlist WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
	`, companyID, ownerType, ownerID)
	return err
}

func (r *pgModelAllowlistRepo) HasAny(ctx context.Context, ownerType, ownerID string) (bool, error) {
	companyID := store.CompanyID(ctx)
	var hasAny bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM model_allowlist
			WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
		)
	`, companyID, ownerType, ownerID).Scan(&hasAny)
	return hasAny, err
}

func (r *pgModelAllowlistRepo) IsAllowed(ctx context.Context, ownerType, ownerID string, modelID int64) (bool, error) {
	companyID := store.CompanyID(ctx)
	var allowed bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM model_allowlist
			WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3 AND model_id = $4
		)
	`, companyID, ownerType, ownerID, modelID).Scan(&allowed)
	return allowed, err
}

func (r *pgModelAllowlistRepo) IsCallTypeAllowed(ctx context.Context, ownerType, ownerID, callType string) (bool, error) {
	companyID := store.CompanyID(ctx)
	var allowed bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM model_allowlist ma
			JOIN models m ON m.model_id = ma.model_id
			WHERE ma.company_id = $1 AND ma.owner_type = $2 AND ma.owner_id = $3
				AND m.type = $4 AND m.enabled = TRUE
		)
	`, companyID, ownerType, ownerID, callType).Scan(&allowed)
	return allowed, err
}

func pruneAllowlistByOwnerIDs(ctx context.Context, db dbQuerier, companyID int64, ownerType string, ownerIDs []string) error {
	if len(ownerIDs) == 0 {
		_, err := db.Exec(ctx, `DELETE FROM model_allowlist WHERE company_id = $1 AND owner_type = $2`, companyID, ownerType)
		return err
	}
	_, err := db.Exec(ctx, `
		DELETE FROM model_allowlist
		WHERE company_id = $1 AND owner_type = $2 AND NOT (owner_id = ANY($3))
	`, companyID, ownerType, ownerIDs)
	return err
}
