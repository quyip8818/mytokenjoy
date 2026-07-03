package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDataSourceTestInvalidCredential422(t *testing.T) {
	app := newTestApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuAuthErrorServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	body := []byte(`{"platform":"feishu","appId":"bad","appSecret":"bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/data-source/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDataSourceImportRetryWithIDs(t *testing.T) {
	var serverURL string
	app := newTestApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		serverURL = server.URL
		cfg.FeishuBaseURL = serverURL
	})
	testutil.ConnectFeishuDataSource(t, &app.Config, app.Store, serverURL)

	body := []byte(`{"ids":["fail-1"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/data-source/import/retry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
}
