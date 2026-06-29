package usage

import (
	"errors"

	"github.com/tokenjoy/backend/internal/domain"
)

func IsNewAPIUnavailable(err error) bool {
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Status == domain.StatusServiceUnavailable
}
