package company

import (
	"strings"

	"github.com/google/uuid"
)

// WalletUsername derives a deterministic NewAPI wallet username from a company UUID.
// Returns the UUID hex without hyphens (32 chars, within NewAPI's 40-char limit).
func WalletUsername(id uuid.UUID) string {
	return strings.ReplaceAll(id.String(), "-", "")
}
