package budget

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

// CombinedKeyEntry is the optional Redis cache row for combined key budget checks.
type CombinedKeyEntry struct {
	Remain float64
	UpdatedAt  time.Time
	Version    int64
}

// CombinedKeyCache is the domain port for combined key budget cache (PG remains authoritative).
type CombinedKeyCache interface {
	Enabled() bool
	Get(ctx context.Context, companyID int64, keyHash string) (CombinedKeyEntry, bool, error)
	Set(ctx context.Context, companyID int64, keyHash string, entry CombinedKeyEntry) error
}

// BlocksCombinedKey reports whether a cached entry should block at the given PG version.
func BlocksCombinedKey(entry CombinedKeyEntry, pgVersion int64) bool {
	if pgVersion <= 0 || entry.Version < pgVersion {
		return false
	}
	return entry.Remain <= 0
}

// RefreshCombinedKeySummaries writes PG-derived summaries into the optional cache.
func RefreshCombinedKeySummaries(
	ctx context.Context,
	cache CombinedKeyCache,
	logger *slog.Logger,
	companyID int64,
	summaries []store.CombinedKeySummary,
) {
	if cache == nil || !cache.Enabled() || len(summaries) == 0 {
		return
	}
	for _, summary := range summaries {
		entry := CombinedKeyEntry{
			Remain: summary.Remain,
			UpdatedAt:  summary.UpdatedAt,
			Version:    summary.Version,
		}
		if err := cache.Set(ctx, companyID, summary.KeyHash, entry); err != nil && logger != nil {
			logger.Warn("gateway budget check set failed", "key_id", summary.PlatformKeyID, "error", err)
		}
	}
}

type noopCombinedKeyCache struct{}

func (noopCombinedKeyCache) Enabled() bool { return false }

func (noopCombinedKeyCache) Get(context.Context, int64, string) (CombinedKeyEntry, bool, error) {
	return CombinedKeyEntry{}, false, nil
}

func (noopCombinedKeyCache) Set(context.Context, int64, string, CombinedKeyEntry) error { return nil }

// NoopCombinedKeyCache is the default when Redis is unavailable or disabled.
var NoopCombinedKeyCache CombinedKeyCache = noopCombinedKeyCache{}

var _ CombinedKeyCache = noopCombinedKeyCache{}
