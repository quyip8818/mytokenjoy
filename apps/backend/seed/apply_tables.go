package seed

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/apply"
)

type TableWriter = apply.TableWriter

func ApplyTables(ctx context.Context, exec TableWriter, snap store.Snapshot) error {
	return apply.ApplyTables(ctx, exec, snap)
}
