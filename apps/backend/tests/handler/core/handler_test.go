package core_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSessionUnauthorized(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cases := []struct {
		name   string
		cookie string
	}{
		{name: "demo missing member id", cookie: ""},
		{name: "invalid token", cookie: "tokenjoy_session_member=missing"},
		{name: "production unauthorized", cookie: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
			if tc.cookie != "" {
				req.Header.Set("Cookie", tc.cookie)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestSessionDemoSuccess(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", testutil.SessionCookie(t, seed.IDMemberAdmin))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload types.SessionContext
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Member.ID != seed.IDMemberAdmin {
		t.Fatalf("expected %s, got %s", seed.IDMemberAdmin, payload.Member.ID)
	}
	if payload.ReadOnly {
		t.Fatal("expected admin session to be writable")
	}
}

func TestUnknownAPIRouteReturnsJSON(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["message"] != "Not found" {
		t.Fatalf("expected Not found message, got %q", body["message"])
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodDelete,
	}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/healthz", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if method == http.MethodHead && rec.Body.Len() > 0 {
				t.Fatalf("expected empty body for HEAD, got %q", rec.Body.String())
			}
			if method != http.MethodHead {
				var payload map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if payload["status"] != "ok" {
					t.Fatalf("expected status ok, got %q", payload["status"])
				}
			}
		})
	}
}

func TestCoreGetEndpoints(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	adminCookie := testhttp.AdminCookie(t)

	t.Run("org roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/org/roles", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var roles []types.Role
		if err := json.NewDecoder(rec.Body).Decode(&roles); err != nil {
			t.Fatal(err)
		}
		if len(roles) != 6 {
			t.Fatalf("expected 6 roles, got %d", len(roles))
		}
		foundSuperAdmin := false
		for _, r := range roles {
			if r.Name == permission.RoleSuperAdmin {
				foundSuperAdmin = true
				break
			}
		}
		if !foundSuperAdmin {
			t.Fatal("expected preset super admin role")
		}
	})

	t.Run("org data source status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/org/data-source/status", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var status types.DataSourceStatus
		if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
			t.Fatal(err)
		}
		if status.Connected {
			t.Fatal("expected disconnected initial status")
		}
	})

	t.Run("budget tree", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/budget/tree", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var tree []types.BudgetNode
		if err := json.NewDecoder(rec.Body).Decode(&tree); err != nil {
			t.Fatal(err)
		}
		if len(tree) == 0 || tree[0].ID != "dept-1" {
			t.Fatal("expected budget tree root dept-1")
		}
	})

	t.Run("keys provider", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/keys/provider", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var providerKeys []types.ProviderKey
		if err := json.NewDecoder(rec.Body).Decode(&providerKeys); err != nil {
			t.Fatal(err)
		}
		if len(providerKeys) < 8 {
			t.Fatalf("expected at least 8 provider keys, got %d", len(providerKeys))
		}
	})

	t.Run("models list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var modelsList []types.ModelInfo
		if err := json.NewDecoder(rec.Body).Decode(&modelsList); err != nil {
			t.Fatal(err)
		}
		if len(modelsList) < 8 {
			t.Fatalf("expected at least 8 models, got %d", len(modelsList))
		}
	})

	t.Run("dashboard cost summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/summary", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var summary types.CostSummary
		if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
			t.Fatal(err)
		}
		if summary.TotalCost < 0 {
			t.Fatal("expected non-negative total cost")
		}
	})

	t.Run("audit operations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit/operations", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var page struct {
			Items []types.OperationLog `json:"items"`
			Total int                  `json:"total"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&page); err != nil {
			t.Fatal(err)
		}
		if page.Total == 0 || len(page.Items) == 0 {
			t.Fatal("expected audit operation logs")
		}
	})
}

func TestKeysPlatformCreateMissingMemberID(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body, _ := json.Marshal(map[string]any{
		"name": "test", "quota": 1000, "modelWhitelist": []string{"gpt-4o"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/keys/platform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBudgetNodeUpdateOversell(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"budget":90000,"reservedPool":1500}`)
	req := httptest.NewRequest(http.MethodPut, "/api/budget/departments/dept-3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}
