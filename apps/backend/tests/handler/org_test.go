package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestMemberCreateHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"name":"测试用户","phone":"13800000000","email":"test@example.com","departmentId":"dept-3"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var member struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&member); err != nil {
		t.Fatal(err)
	}
	if member.Name != "测试用户" {
		t.Fatalf("expected name 测试用户, got %q", member.Name)
	}
}

func TestBatchImportHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"rows":[{"name":"导入用户","phone":"13900000000","email":"import@example.com","departmentName":"后端组"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/members/batch-import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result types.MemberBatchImportResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected imported=1, got %d", result.Imported)
	}
	if len(result.Failures) != 0 {
		t.Fatalf("expected no failures, got %v", result.Failures)
	}
}
