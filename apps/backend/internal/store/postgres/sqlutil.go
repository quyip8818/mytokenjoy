package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

func pruneByID(ctx context.Context, db dbQuerier, table string, ids []uuid.UUID) error {
	if len(ids) == 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s`, table))
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE NOT (id = ANY($1))`, table), ids); err != nil {
		return fmt.Errorf("prune %s: %w", table, err)
	}
	return nil
}

func pruneByIDForCompany(ctx context.Context, db dbQuerier, table string, companyID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE company_id = $1`, table), companyID)
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE company_id = $1 AND NOT (id = ANY($2))`, table), companyID, ids); err != nil {
		return fmt.Errorf("prune %s: %w", table, err)
	}
	return nil
}

func pruneByColumnForCompany(ctx context.Context, db dbQuerier, table, column string, companyID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE company_id = $1`, table), companyID)
		return err
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE company_id = $1 AND NOT (%s = ANY($2))`, table, column)
	if _, err := db.Exec(ctx, query, companyID, ids); err != nil {
		return fmt.Errorf("prune %s: %w", table, err)
	}
	return nil
}
