package testutil

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
)

const SessionCookieAdmin = "tokenjoy_session_member=m-admin"

func SessionCookie(memberID string) string {
	return "tokenjoy_session_member=" + memberID
}

func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	cfg := TestConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return app.New(cfg, logger)
}

func NewTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewTestApp(t, nil).Router
}
