package company

import "github.com/tokenjoy/backend/internal/store"

func IsRelayBlocked(status string) bool {
	return status != store.CompanyStatusActive
}
