package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDepartmentUpdateHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"name":"Updated Team"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/org/departments/dept-5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentDeleteLeafHTTP(t *testing.T) {
	router := newTestRouter(t)
	createBody := []byte(`{"name":"Temp Leaf","parentId":"dept-2"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/org/departments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Cookie", adminSessionCookie(t))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create expected 200, got %d", createRec.Code)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/org/departments/"+created.ID, nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMembersTransferHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"memberIds":["` + seed.IDMember1 + `"],"departmentId":"dept-4"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/members/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDataSourceUpdateSuccessHTTP(t *testing.T) {
	app := newTestApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	body := []byte(`{"platform":"feishu","appId":"cli_test","appSecret":"secret_test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/org/data-source", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDataSourceImportHTTP(t *testing.T) {
	app := newTestApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	testutil.ConnectFeishuDataSource(t, &app.Config, app.Store, app.Config.FeishuBaseURL)
	req := httptest.NewRequest(http.MethodPost, "/api/org/data-source/import", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSyncTriggerWritesLogHTTP(t *testing.T) {
	app := newTestApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	testutil.ConnectFeishuDataSource(t, &app.Config, app.Store, app.Config.FeishuBaseURL)
	ctx := testutil.Ctx()
	syncLogs, err := app.Store.Org().SyncLogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before := len(syncLogs)
	req := httptest.NewRequest(http.MethodPost, "/api/org/sync/trigger", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	syncLogs, err = app.Store.Org().SyncLogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(syncLogs) <= before {
		t.Fatal("expected sync log after trigger")
	}
}
