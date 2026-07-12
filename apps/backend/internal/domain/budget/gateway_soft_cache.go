package budget

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

// GatewaySoftEntry is the optional Redis cache row for Gateway soft budget checks.
type GatewaySoftEntry struct {
	SoftRemain float64
	UpdatedAt  time.Time
	Version    int64
}

// GatewaySoftCache is the domain port for Gateway soft-block cache (PG remains authoritative).
type GatewaySoftCache interface {
	Enabled() bool
	Get(ctx context.Context, companyID int64, keyHash string) (GatewaySoftEntry, bool, error)
	Set(ctx context.Context, companyID int64, keyHash string, entry GatewaySoftEntry) error
}

// BlocksGatewaySoft reports whether a cached entry should block at the given PG version.
func BlocksGatewaySoft(entry GatewaySoftEntry, pgVersion int64) bool {
	if pgVersion <= 0 || entry.Version < pgVersion {
		return false
	}
	return entry.SoftRemain <= 0
}

// RefreshGatewaySoftSummaries writes PG-derived summaries into the optional cache.
func RefreshGatewaySoftSummaries(
	ctx context.Context,
	cache GatewaySoftCache,
	logger *slog.Logger,
	companyID int64,
	summaries []store.GatewaySoftSummary,
) {
	if cache == nil || !cache.Enabled() || len(summaries) == 0 {
		return
	}
	for _, summary := range summaries {
		entry := GatewaySoftEntry{
			SoftRemain: summary.SoftRemain,
			UpdatedAt:  summary.UpdatedAt,
			Version:    summary.Version,
		}
		if err := cache.Set(ctx, companyID, summary.KeyHash, entry); err != nil && logger != nil {
			logger.Warn("gateway budget check set failed", "key_id", summary.PlatformKeyID, "error", err)
		}
	}
}

type noopGatewaySoftCache struct{}

func (noopGatewaySoftCache) Enabled() bool { return false }

func (noopGatewaySoftCache) Get(context.Context, int64, string) (GatewaySoftEntry, bool, error) {
	return GatewaySoftEntry{}, false, nil
}

func (noopGatewaySoftCache) Set(context.Context, int64, string, GatewaySoftEntry) error { return nil }

// NoopGatewaySoftCache is the default when Redis is unavailable or disabled.
var NoopGatewaySoftCache GatewaySoftCache = noopGatewaySoftCache{}

var _ GatewaySoftCache = noopGatewaySoftCache{}
