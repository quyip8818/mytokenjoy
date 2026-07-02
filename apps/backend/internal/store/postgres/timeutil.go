package postgres

import (
	"time"

	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
)

func parseAPITime(value string) (time.Time, error) {
	return pkgtime.Parse(value)
}

func formatSyncLogTime(t time.Time) string {
	return pkgtime.FormatSyncLog(t)
}

func formatDateOnly(t time.Time) string {
	return pkgtime.FormatDateOnly(t)
}

func parseTimeOrNow(value string) (time.Time, error) {
	return pkgtime.ParseOrNow(value)
}
