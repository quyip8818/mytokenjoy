package types

import "github.com/google/uuid"

// PrecheckCacheInvalidator is the shared interface for invalidating gateway precheck
// cache entries. Defined in the shared kernel so both keys and company domains can
// depend on it without importing gateway (which would create circular imports).
type PrecheckCacheInvalidator interface {
	InvalidateByKeyID(platformKeyID uuid.UUID)
	InvalidateCompany(companyID uuid.UUID)
}

// NoopPrecheckCacheInvalidator does nothing (used when gateway is disabled or in tests).
type NoopPrecheckCacheInvalidator struct{}

func (NoopPrecheckCacheInvalidator) InvalidateByKeyID(uuid.UUID) {}
func (NoopPrecheckCacheInvalidator) InvalidateCompany(uuid.UUID) {}
