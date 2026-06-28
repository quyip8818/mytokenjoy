package store

import "context"

func (m *Memory) WithTx(ctx context.Context, fn func(Store) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(m)
}
