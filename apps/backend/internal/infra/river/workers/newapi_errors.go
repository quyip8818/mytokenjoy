package workers

import (
	"strings"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
)

// IsNonRetryableNewAPIError reports permanent upstream/client errors that must not
// burn River retries (e.g. PG bigint overflow on quota writes).
func IsNonRetryableNewAPIError(err error) bool {
	if err == nil {
		return false
	}
	if outbox.IsPermanentOutboxError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "bigint out of range") ||
		strings.Contains(msg, "sqlstate 22003") ||
		strings.Contains(msg, "topup quota delta out of range") ||
		strings.Contains(msg, "newapi wallet user id required")
}

func cancelIfNonRetryable(err error) error {
	if IsNonRetryableNewAPIError(err) {
		return river.JobCancel(err)
	}
	return err
}
