package company

import (
	"strings"

	"github.com/google/uuid"
)

// WalletUsername derives a deterministic NewAPI wallet username from a company UUID.
// Returns "w_" + first 16 hex digits (18 chars total, within NewAPI's 20-char default limit).
func WalletUsername(id uuid.UUID) string {
	hex := strings.ReplaceAll(id.String(), "-", "")
	return "w_" + hex[:16]
}
