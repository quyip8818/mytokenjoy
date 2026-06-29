package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func (m *Store) WithTx(ctx context.Context, fn func(store.Store) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(m)
}
