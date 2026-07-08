//go:build testhook

package testutil

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/runtime"
)

func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	cfg := TestConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	storeCfg := cfg
	if cfg.IsProdProfile() {
		storeCfg.Profile = config.ProfileDemo
	}
	_, st := NewTestStore(t, func(c *config.Config) { *c = storeCfg })
	if storeCfg.IsDemoProfile() {
		ctx := Ctx()
		if err := runtime.ApplyUsageBuckets(ctx, st, storeCfg); err != nil {
			t.Fatalf("apply usage buckets: %v", err)
		}
		if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
			t.Fatalf("apply recharge orders: %v", err)
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	application, err := app.NewWithStore(cfg, logger, st, app.WithoutWorker())
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	return application
}

func NewTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewTestApp(t, nil).Router
}
