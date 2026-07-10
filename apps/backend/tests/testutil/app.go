//go:build testhook

package testutil

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
)

func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	cfg := TestConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	_, st := NewTestStore(t, WithConfig(cfg))
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
