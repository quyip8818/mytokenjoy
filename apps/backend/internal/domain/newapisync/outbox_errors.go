package newapisync

import (
	"strings"

	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func IsPermanentOutboxError(err error) bool {
	if err == nil {
		return false
	}
	if domainusage.IsNewAPIUnavailable(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "unmarshal") || strings.Contains(msg, "unknown newapi sync outbox kind")
}
