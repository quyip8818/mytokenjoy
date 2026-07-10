package company

import "github.com/tokenjoy/backend/internal/store"

func IsGatewayBlocked(status string) bool {
	return status != store.CompanyStatusActive
}
