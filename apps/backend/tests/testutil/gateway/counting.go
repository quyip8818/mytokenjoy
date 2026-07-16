//go:build testhook

package gatewayfix

import (
	"context"
	"sync/atomic"
	"testing"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/store"
)

// CountingGatewayPrecheck wraps a GatewayPrecheckRepository and records LoadPrecheckContext calls.
type CountingGatewayPrecheck struct {
	inner store.GatewayPrecheckRepository
	calls atomic.Int32
}

func NewCountingGatewayPrecheck(inner store.GatewayPrecheckRepository) *CountingGatewayPrecheck {
	return &CountingGatewayPrecheck{inner: inner}
}

func (c *CountingGatewayPrecheck) LoadPrecheckContext(ctx context.Context, keyHash string) (*store.PrecheckContextRow, error) {
	c.calls.Add(1)
	return c.inner.LoadPrecheckContext(ctx, keyHash)
}

func (c *CountingGatewayPrecheck) Calls() int32 {
	return c.calls.Load()
}

func BuildGatewayWithPrecheckLoader(t *testing.T, scenario GatewayScenario, loader store.GatewayPrecheckRepository) domaingateway.GatewayService {
	t.Helper()
	precheck := domaingateway.NewPrecheckServiceLegacy(loader, scenario.Cfg.Clock(), nil)
	gw, err := domaingateway.NewGatewayService(scenario.Cfg, precheck, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return gw
}

func BuildGatewayWithCountingPrecheck(t *testing.T, scenario GatewayScenario) (domaingateway.GatewayService, *CountingGatewayPrecheck) {
	t.Helper()
	counter := NewCountingGatewayPrecheck(scenario.Store.GatewayPrecheck())
	return BuildGatewayWithPrecheckLoader(t, scenario, counter), counter
}
