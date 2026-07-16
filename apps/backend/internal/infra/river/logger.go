package riverinfra

import (
	"context"
	"log/slog"
	"strings"
)

const producerJobCountsMsg = "Producer job counts"

// QuietLogger drops River producer heartbeat lines (queue counters) while
// keeping Warn/Error and real job failure logs. Stuck-count heartbeats are kept.
func QuietLogger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return slog.New(&heartbeatFilter{inner: logger.Handler()})
}

type heartbeatFilter struct {
	inner slog.Handler
}

func (h *heartbeatFilter) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *heartbeatFilter) Handle(ctx context.Context, r slog.Record) error {
	if isProducerHeartbeat(r) {
		return nil
	}
	return h.inner.Handle(ctx, r)
}

func (h *heartbeatFilter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &heartbeatFilter{inner: h.inner.WithAttrs(attrs)}
}

func (h *heartbeatFilter) WithGroup(name string) slog.Handler {
	return &heartbeatFilter{inner: h.inner.WithGroup(name)}
}

func isProducerHeartbeat(r slog.Record) bool {
	if r.Level > slog.LevelInfo {
		return false
	}
	if !strings.Contains(r.Message, producerJobCountsMsg) {
		return false
	}
	stuck := 0
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "num_jobs_stuck" {
			switch a.Value.Kind() {
			case slog.KindInt64:
				stuck = int(a.Value.Int64())
			case slog.KindUint64:
				stuck = int(a.Value.Uint64())
			}
		}
		return true
	})
	return stuck == 0
}
