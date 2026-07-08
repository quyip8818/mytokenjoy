package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelAllowlistRepo struct {
	db dbQuerier
}

func (r *pgModelAllowlistRepo) List(ctx context.Context, ownerType, ownerID string) ([]string, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT m.name
		FROM model_allowlist ma
		JOIN models m ON m.company_id = ma.company_id AND m.id = ma.model_id
		WHERE ma.company_id = $1 AND ma.owner_type = $2 AND ma.owner_id = $3
		ORDER BY m.name
	`, companyID, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	models := make([]string, 0)
	for rows.Next() {
		var modelName string
		if err := rows.Scan(&modelName); err != nil {
			return nil, err
		}
		models = append(models, modelName)
	}
	return models, rows.Err()
}

func (r *pgModelAllowlistRepo) Replace(ctx context.Context, ownerType, ownerID string, models []string) error {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, `
		DELETE FROM model_allowlist WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
	`, companyID, ownerType, ownerID); err != nil {
		return fmt.Errorf("clear allowlist: %w", err)
	}
	for _, modelName := range models {
		var modelID string
		err := r.db.QueryRow(ctx, `
			SELECT id FROM models WHERE company_id = $1 AND name = $2
		`, companyID, modelName).Scan(&modelID)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("unknown model %q", modelName)
		}
		if err != nil {
			return fmt.Errorf("resolve model %q: %w", modelName, err)
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO model_allowlist (company_id, owner_type, owner_id, model_id)
			VALUES ($1, $2, $3, $4)
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

func (r *pgModelAllowlistRepo) IsAllowed(ctx context.Context, ownerType, ownerID, modelName string) (bool, error) {
	companyID := store.CompanyID(ctx)
	var allowed bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM model_allowlist ma
			JOIN models m ON m.company_id = ma.company_id AND m.id = ma.model_id
			WHERE ma.company_id = $1 AND ma.owner_type = $2 AND ma.owner_id = $3 AND m.name = $4
		)
	`, companyID, ownerType, ownerID, modelName).Scan(&allowed)
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
