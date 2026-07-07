package budget_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestBudgetApprovalsList(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/budget/approvals", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var items []types.BudgetApproval
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 {
		t.Fatalf("expected seeded approvals, got %d", len(items))
	}
}

func TestBudgetApprovalResolve(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"status":"approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/budget/approvals/appr-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var item types.BudgetApproval
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.Status != "approved" {
		t.Fatalf("expected approved, got %s", item.Status)
	}
}
