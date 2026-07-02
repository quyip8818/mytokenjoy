package seed

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

func insertCompany(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	t := snap.Company
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, t.ID, t.Slug, t.Name, t.Status); err != nil {
		return fmt.Errorf("seed company: %w", err)
	}
	return nil
}
