package newapisync

import (
	"github.com/tokenjoy/backend/internal/domain"
)

// IsPermanentOutboxError reports errors that should not be retried by the outbox worker.
// ServiceUnavailable is permanent here because domain code uses it for disabled NewAPI,
// not transient upstream outages (those surface as generic errors and retry).
func IsPermanentOutboxError(err error) bool {
	if err == nil {
		return false
	}
	return domain.IsServiceUnavailable(err)
}
