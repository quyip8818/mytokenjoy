package company

import "github.com/tokenjoy/backend/internal/store"

// IsTestingAccount returns true for non-production account types that allow
// test-model access, simulated consumption, and mock quota grants.
// ponytail: single source of truth for "demo|trial|testing" checks across the codebase.
func IsTestingAccount(companyType string) bool {
	switch companyType {
	case store.CompanyTypeTrial, store.CompanyTypeDemo, store.CompanyTypeTesting:
		return true
	default:
		return false
	}
}
