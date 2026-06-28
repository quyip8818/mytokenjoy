package testutil

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
)

func AssertDomainStatus(t *testing.T, err error, status int) {
	t.Helper()
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain error status %d, got %v", status, err)
	}
	if domainErr.Status != status {
		t.Fatalf("expected status %d, got %d message=%q", status, domainErr.Status, domainErr.Message)
	}
}

func StrPtr(v string) *string {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}
