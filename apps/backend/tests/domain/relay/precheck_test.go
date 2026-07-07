package relay_test

import (
	"testing"

	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func TestPrecheckRejectsZeroBudget(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := relayfix.ConfigureGatewayStore(t, st, relayfix.GatewayScenarioOpts{Budget: 0})

	mapping, err := st.Relay().GetMappingByFullKey(ctx, fullKey)
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
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

	mapping, err := st.Relay().GetMappingByFullKey(ctx, fullKey)
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
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

	mapping, err := st.Relay().GetMappingByFullKey(ctx, fullKey)
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
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

	mapping, err := st.Relay().GetMappingByFullKey(ctx, fullKey)
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	company.Status = "suspended"

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err == nil {
		t.Fatal("expected suspended company error")
	}
}
