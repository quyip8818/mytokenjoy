package authz_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/config"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAuthzWriteEndpoints(t *testing.T) {
	router := testhttp.NewRouter(t)
	for _, tc := range authzWriteCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Reader
			if tc.body != "" {
				body = bytes.NewReader([]byte(tc.body))
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Content-Type", "application/json")
			if tc.cookie != "" {
				req.Header.Set("Cookie", tc.cookie)
			}
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d body=%s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestProdGetReadForbiddenForMember(t *testing.T) {
	router := testhttp.NewProdRouter(t)
	for _, tc := range prodGetForbiddenCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.cookie != "" {
				req.Header.Set("Cookie", tc.cookie)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d body=%s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSyncTriggerWithAPIKey(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.SyncTriggerAPIKey = "test-sync-key"
		cfg.FeishuBaseURL = env.ServerURL
	})
	testutil.ConnectFeishuDataSource(t, &app.Config, app.Store, env.ServerURL)

	req := httptest.NewRequest(http.MethodPost, "/api/org/sync/trigger", nil)
	req.Header.Set(httpmiddleware.SyncTriggerAPIKeyHeader, "test-sync-key")
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
