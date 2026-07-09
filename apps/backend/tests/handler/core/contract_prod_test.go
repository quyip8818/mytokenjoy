package core_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func TestGetContractRequiresSession(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	for _, tc := range getContractCases() {
		if tc.path == "/healthz" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetContractWithAdminCookie(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	for _, tc := range getContractCases() {
		if tc.path == "/healthz" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Cookie", testhttp.AdminCookie(t))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}
			var payload any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("expected JSON body: %v", err)
			}
		})
	}
}
