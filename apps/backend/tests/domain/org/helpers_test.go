package org_test

import (
	"errors"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/store"
)

func newTestOrgService(t *testing.T) org.Service {
	t.Helper()
	_, _, svc := orgfix.NewServiceFromStore(t)
	return svc
}

func newTestOrgServiceWithStore(t *testing.T) (org.Service, store.Store) {
	t.Helper()
	_, st, svc := orgfix.NewServiceFromStore(t)
	return svc, st
}

func asDomainError(t *testing.T, err error) *domain.DomainError {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected domain error, got %v", err)
	}
	return de
}
