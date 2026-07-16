package gateway

// KeyCacheInvalidator is the interface exposed to other domain services
// for invalidating cached precheck data when keys or companies are modified.
type KeyCacheInvalidator interface {
	InvalidateByKeyID(platformKeyID string)
	InvalidateCompany(companyID int64)
}

// Verify PrecheckCache satisfies the interface.
var _ KeyCacheInvalidator = (*PrecheckCache)(nil)

// NoopInvalidator does nothing (used when gateway is disabled or in tests).
type NoopInvalidator struct{}

func (NoopInvalidator) InvalidateByKeyID(string) {}
func (NoopInvalidator) InvalidateCompany(int64)  {}
