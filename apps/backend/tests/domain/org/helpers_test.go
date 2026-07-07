package org_test

import (
	"errors"
	"log/slog"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
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

func newTestService(t *testing.T) org.Service {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	factory := datasource.NewFactory(cfg)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil, nil, relay.NewChannelPolicy(cfg))
	notifier := notification.NewService(cfg, st, slog.Default())
	return org.NewService(cfg, st, factory, lifecycle, notifier, common.NewDelayer(false), slog.Default())
}

func asDomainError(t *testing.T, err error) *domain.DomainError {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected domain error, got %v", err)
	}
	return de
}
