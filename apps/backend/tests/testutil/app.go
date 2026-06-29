//go:build testhook

package testutil

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/seed"
)

func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	cfg := TestConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	st := NewMemoryStore(t, cfg)
	if cfg.IsDemoProfile() {
		ctx := context.Background()
		if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
			t.Fatalf("apply usage buckets: %v", err)
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
