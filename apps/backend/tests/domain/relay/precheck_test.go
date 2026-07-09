package relay_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func TestPrecheckRejectsZeroBudget(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 0})

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err == nil {
		t.Fatal("expected budget exceeded error")
	}
}

func TestPrecheckRejectsInactivePlatformKey(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 1000})

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		keys[i].Status = "disabled"
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err == nil {
		t.Fatal("expected inactive platform key error")
	}
}

func TestPrecheckRejectsModelNotInWhitelist(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "unknown-model",
	})
	if err == nil {
		t.Fatal("expected model not allowed error")
	}
}

func TestPrecheckRejectsSuspendedCompany(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	company.Status = "suspended"

	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err == nil {
		t.Fatal("expected suspended company error")
	}
}

func TestPrecheckRejectsPendingWalletSyncWithDrift(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Company().UpdateWalletPoint(ctx, company.ID, 100000, nil); err != nil {
		t.Fatal(err)
	}
	company.BalancePoint = 100000
	if err := st.Relay().EnqueueWalletSync(ctx, company.ID); err != nil {
		t.Fatal(err)
	}

	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(newapi.ToNewAPIUnits(1, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
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

func TestPrecheckAllowsPendingWalletSyncWithoutDrift(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 1000})

	mapping, err := st.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(fullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	balancePoint := 100000.0
	if err := st.Company().UpdateWalletPoint(ctx, company.ID, balancePoint, nil); err != nil {
		t.Fatal(err)
	}
	company.BalancePoint = balancePoint
	if err := st.Relay().EnqueueWalletSync(ctx, company.ID); err != nil {
		t.Fatal(err)
	}

	models, err := st.Models().Models(ctx)
	if err != nil {
		t.Fatal(err)
	}
	naUnits := newapi.ToQuotaUnits(balancePoint, models, nil)
	precheck := relayfix.NewPrecheckService(st, relayfix.NewStubWallet(naUnits))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err != nil {
		t.Fatalf("expected precheck to pass with aligned wallet, got %v", err)
	}
}
