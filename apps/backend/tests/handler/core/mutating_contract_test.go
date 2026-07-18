package core_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestMutatingContractEndpoints(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	t.Run("budget department update", func(t *testing.T) {
		// dept-6 has no demo oversubscription; dept-3 is reserved for overrun scenarios.
		deptID := contract.IDDept6.String()
		const wantBudget = 21000000
		body := []byte(`{"budget":21000000}`)
		req := httptest.NewRequest(http.MethodPut, "/api/budget/departments/"+deptID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		var node types.BudgetNode
		if err := json.NewDecoder(rec.Body).Decode(&node); err != nil {
			t.Fatal(err)
		}
		if node.Budget != wantBudget {
			t.Fatalf("expected budget %v, got %v", wantBudget, node.Budget)
		}
	})

	t.Run("budget approval reject", func(t *testing.T) {
		body := []byte(`{"status":"rejected","rejectReason":"test"}`)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/budget/approvals/%s", contract.IDBudgetApproval1), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		var item types.BudgetApproval
		if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
			t.Fatal(err)
		}
		if item.Status != "rejected" {
			t.Fatalf("expected rejected, got %s", item.Status)
		}
		if item.RejectReason == nil || *item.RejectReason != "test" {
			t.Fatalf("expected reject reason test, got %v", item.RejectReason)
		}
	})
}
