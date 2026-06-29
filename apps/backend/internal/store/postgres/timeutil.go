package postgres

import (
	"time"

	"github.com/tokenjoy/backend/internal/store/timeparse"
)

func parseAPITime(value string) (time.Time, error) {
	return timeparse.Parse(value)
}

func formatSyncLogTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func formatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}
