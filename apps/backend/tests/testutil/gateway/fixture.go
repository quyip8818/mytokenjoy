//go:build testhook

package gatewayfix

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

// PrecheckFixture wires store seeding and domain precheck for gateway unit tests.
type PrecheckFixture struct {
	Cfg      config.Config
	Store    store.Store
	Ctx      context.Context
	FullKey  string
	Precheck *domaingateway.PrecheckService
}

func NewPrecheckFixture(t *testing.T, opts GatewayScenarioOpts, cfgOpts ...testutil.ConfigOption) PrecheckFixture {
	t.Helper()
	allOpts := append([]testutil.ConfigOption{testutil.WithNewAPIEnabled(true)}, cfgOpts...)
	cfg, st := testutil.NewTestStore(t, allOpts...)
	fullKey := ConfigureGatewayStore(t, cfg, st, opts)
	return PrecheckFixture{
		Cfg:      cfg,
		Store:    st,
		Ctx:      testutil.Ctx(),
		FullKey:  fullKey,
		Precheck: NewPrecheckService(cfg, st, nil),
	}
}

func (f PrecheckFixture) KeyHash() string {
	return store.HashPlatformKey(f.FullKey)
}

func (f PrecheckFixture) Run(model string, skipModelCheck bool) error {
	_, err := f.Precheck.Run(f.Ctx, f.KeyHash(), model, domaingateway.PrecheckOpts{SkipModelCheck: skipModelCheck})
	return err
}

func (f PrecheckFixture) LoadPrecheckRow(t *testing.T) *store.PrecheckContextRow {
	t.Helper()
	row, err := f.Store.GatewayPrecheck().LoadPrecheckContext(f.Ctx, f.KeyHash())
	if err != nil {
		t.Fatal(err)
	}
	return row
}
