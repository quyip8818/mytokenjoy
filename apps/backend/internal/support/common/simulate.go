package common

import (
	"context"
	"time"
)

type Delayer struct {
	Enabled bool
}

func NewDelayer(enabled bool) Delayer {
	return Delayer{Enabled: enabled}
}

func (d Delayer) Wait(ctx context.Context, duration time.Duration) error {
	if !d.Enabled {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
