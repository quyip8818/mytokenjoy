package handler_test

import (
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

const sessionCookie = testutil.SessionCookieAdmin

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return testutil.NewTestRouter(t)
}

func newTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	return testutil.NewTestApp(t, mutate)
}
