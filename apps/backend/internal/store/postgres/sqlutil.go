package postgres

import (
	"context"
	"fmt"
)

func pruneByID(ctx context.Context, db dbQuerier, table string, ids []string) error {
	if len(ids) == 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s`, table))
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE NOT (id = ANY($1))`, table), ids); err != nil {
		return fmt.Errorf("prune %s: %w", table, err)
	}
	return nil
}

func pruneByColumn(ctx context.Context, db dbQuerier, table, column string, ids []string) error {
	if len(ids) == 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s`, table))
		return err
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE NOT (%s = ANY($1))`, table, column)
	if _, err := db.Exec(ctx, query, ids); err != nil {
		return fmt.Errorf("prune %s: %w", table, err)
	}
	return nil
}
