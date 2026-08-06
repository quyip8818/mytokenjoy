package core_test

import (
	"bytes"
	"encoding/json"
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
		// Use a value high enough to always exceed current budget (no-decrease rule).
		deptID := contract.IDDept6.String()
		const wantBudget = 20000
		body := []byte(`{"budget":20000}`)
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
}
