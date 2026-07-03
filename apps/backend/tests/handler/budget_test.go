package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestUpdateNodeHTTPSuccess(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"budget":21000,"reservedPool":1500}`)
	req := httptest.NewRequest(http.MethodPut, "/api/budget/departments/dept-3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var node types.BudgetNode
	if err := json.NewDecoder(rec.Body).Decode(&node); err != nil {
		t.Fatal(err)
	}
	if node.Budget != 21000 {
		t.Fatalf("expected budget 21000, got %v", node.Budget)
	}
}
