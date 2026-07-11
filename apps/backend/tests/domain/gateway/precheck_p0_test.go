package gateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

type failingWalletSyncQueue struct{}

func (failingWalletSyncQueue) EnqueueWalletSync(context.Context, int64) error {
	panic("unexpected EnqueueWalletSync")
}
func (failingWalletSyncQueue) ClaimPendingWalletSync(context.Context, int) ([]store.WalletSyncQueueEntry, error) {
	panic("unexpected ClaimPendingWalletSync")
}
func (failingWalletSyncQueue) MarkWalletSyncDone(context.Context, string) error {
	panic("unexpected MarkWalletSyncDone")
}
func (failingWalletSyncQueue) HasPendingWalletSync(context.Context, int64) (bool, error) {
	return false, errors.New("wallet sync lookup failed")
}

func TestPrecheckRejectsWalletSyncLookupError(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := gatewaytf.ConfigureGatewayStore(t, cfg, st, gatewaytf.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.PlatformKeyMappings().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected platform key mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if company.NewAPIWalletUserID == nil {
		t.Fatal("expected newapi wallet user")
	}

	precheck := domaingateway.NewPrecheckService(
		st.BudgetSnapshots(),
		st.Org().Nodes(),
		st.Budget(),
		st.Org(),
		st.Keys(),
		st.Models(),
		gatewaytf.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)),
		failingWalletSyncQueue{},
		cfg.Clock(),
	)
	err = precheck.Run(ctx, domaingateway.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.RetryAfter == nil {
		t.Fatalf("expected retry-after domain error, got %v", err)
	}
	if *domainErr.RetryAfter != common.WalletSyncRetryAfterSecs {
		t.Fatalf("expected retry-after %d, got %d", common.WalletSyncRetryAfterSecs, *domainErr.RetryAfter)
	}
}

func TestPrecheckAllowsModelsListingWithoutModelField(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := gatewaytf.ConfigureGatewayStore(t, cfg, st, gatewaytf.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.PlatformKeyMappings().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected platform key mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := gatewaytf.NewPrecheckService(cfg, st, gatewaytf.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domaingateway.PrecheckInput{
		Mapping:        mapping,
		Company:        company,
		SkipModelCheck: true,
	})
	if err != nil {
		t.Fatalf("expected models listing precheck to pass, got %v", err)
	}
}

func TestGatewayAllowsModelsListing(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:      1000,
		WalletQuota: 10000,
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+scenario.FullKey)

	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /v1/models, got %d body=%s", rec.Code, rec.Body.String())
	}
}
