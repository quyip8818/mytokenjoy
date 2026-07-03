package testhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/saas"
)

func NewRouter(t *testing.T) http.Handler {
	t.Helper()
	return testutil.NewTestRouter(t)
}

func NewApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	return testutil.NewTestApp(t, mutate)
}

func AdminCookie(t *testing.T) string {
	t.Helper()
	return testutil.SessionCookieAdmin(t)
}

func SaaSRouter(t *testing.T, mock *saas.NewAPIMock) http.Handler {
	t.Helper()
	return saas.NewRouter(t, mock)
}

func NewProdRouter(t *testing.T) http.Handler {
	t.Helper()
	return testutil.NewTestApp(t, func(cfg *config.Config) {
		testutil.WithProfile(config.ProfileProd)(cfg)
	}).Router
}

func ServeAuthz(t *testing.T, router http.Handler, method, path, cookie, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
