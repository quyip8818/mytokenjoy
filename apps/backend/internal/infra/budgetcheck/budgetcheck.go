// Package budgetcheck is the optional combined key budget Redis cache.
// PG combined_key_remain is authoritative; Redis only enhances when version >= PG.
package budgetcheck

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type Entry struct {
	Remain    float64   `json:"remain"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int64     `json:"version"`
}

type Store interface {
	Enabled() bool
	Get(ctx context.Context, companyID int64, keyHash string) (Entry, bool, error)
	Set(ctx context.Context, companyID int64, keyHash string, entry Entry) error
}

func BlocksWithVersion(entry Entry, pgVersion int64) bool {
	if pgVersion <= 0 || entry.Version < pgVersion {
		return false
	}
	return entry.Remain <= 0
}

func Key(companyID int64, keyHash string) string {
	return fmt.Sprintf("gateway:budget_check:%d:%s", companyID, keyHash)
}

func RefreshSummaries(
	ctx context.Context,
	cache Store,
	logger *slog.Logger,
	companyID int64,
	summaries []store.CombinedKeySummary,
) {
	if cache == nil || !cache.Enabled() || len(summaries) == 0 {
		return
	}
	for _, summary := range summaries {
		entry := Entry{
			Remain:    summary.Remain,
			UpdatedAt: summary.UpdatedAt,
			Version:   summary.Version,
		}
		if err := cache.Set(ctx, companyID, summary.KeyHash, entry); err != nil && logger != nil {
			logger.Warn("gateway budget check set failed", "key_id", summary.PlatformKeyID, "error", err)
		}
	}
}

type Noop struct{}

func (Noop) Enabled() bool { return false }

func (Noop) Get(context.Context, int64, string) (Entry, bool, error) {
	return Entry{}, false, nil
}

func (Noop) Set(context.Context, int64, string, Entry) error { return nil }

var _ Store = Noop{}
