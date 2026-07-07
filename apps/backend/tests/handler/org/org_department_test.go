package org_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDepartmentUpdateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"name":"Updated Team"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/org/departments/dept-5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentDeleteLeafHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	createBody := []byte(`{"name":"Temp Leaf","parentId":"dept-2"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/org/departments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Cookie", testhttp.AdminCookie(t))
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
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMembersTransferHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"memberIds":["` + seed.IDMember1 + `"],"departmentId":"dept-4"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/members/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDataSourceUpdateSuccessHTTP(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	body := []byte(`{"platform":"feishu","appId":"cli_test","appSecret":"secret_test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/org/data-source", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDataSourceImportHTTP(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		server := testutil.StartFeishuMockServer(t)
		cfg.FeishuBaseURL = server.URL
	})
	testutil.ConnectFeishuDataSource(t, &app.Config, app.Store, app.Config.FeishuBaseURL)
	req := httptest.NewRequest(http.MethodPost, "/api/org/data-source/import", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSyncTriggerWritesLogHTTP(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, func(cfg *config.Config) {
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
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
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

func TestSyncConfigUpdateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"enabled":true,"startTime":"01:00","frequencyHours":8,"deleteMemberThreshold":5,"deleteDepartmentThreshold":3}`)
	putReq := httptest.NewRequest(http.MethodPut, "/api/org/sync/config", bytes.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Cookie", testhttp.AdminCookie(t))
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("update config: expected 200, got %d body=%s", putRec.Code, putRec.Body.String())
	}
	getReq := httptest.NewRequest(http.MethodGet, "/api/org/sync/config", nil)
	getReq.Header.Set("Cookie", testhttp.AdminCookie(t))
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get config: expected 200, got %d body=%s", getRec.Code, getRec.Body.String())
	}
	var cfg struct {
		Enabled        bool   `json:"enabled"`
		StartTime      string `json:"startTime"`
		FrequencyHours int    `json:"frequencyHours"`
	}
	if err := json.NewDecoder(getRec.Body).Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.StartTime != "01:00" || cfg.FrequencyHours != 8 {
		t.Fatalf("unexpected config %+v", cfg)
	}
}
