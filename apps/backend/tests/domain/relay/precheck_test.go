package relay_test

import (
	"context"
	"testing"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestPrecheckRejectsZeroBudget(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := testutil.ConfigureGatewayStore(t, st, testutil.GatewayScenarioOpts{Budget: 0})

	mapping, err := st.Relay().GetMappingByFullKey(ctx, fullKey)
	if err != nil || mapping == nil {
		t.Fatal("expected relay mapping")
	}
	company, err := st.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), &stubWallet{quota: newapi.ToNewAPIUnits(100, nil, nil)})
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
	_, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()
	fullKey := testutil.ConfigureGatewayStore(t, st, testutil.GatewayScenarioOpts{Budget: 1000})

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

	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), &stubWallet{quota: newapi.ToNewAPIUnits(100, nil, nil)})
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err == nil {
		t.Fatal("expected inactive platform key error")
	}
}

type stubWallet struct {
	quota int64
}

func (s *stubWallet) AvailableQuota(_ context.Context, _ int64) (int64, error) {
	return s.quota, nil
}

var _ domaincompany.WalletService = (*stubWallet)(nil)
