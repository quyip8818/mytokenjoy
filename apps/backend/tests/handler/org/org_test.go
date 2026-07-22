package org_test

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

func TestMemberCreateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(fmt.Sprintf(`{"name":"测试用户","phone":"13800000000","email":"test@example.com","departmentId":"%s"}`, contract.IDDept3.String()))
	req := httptest.NewRequest(http.MethodPost, "/api/org/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var member struct {
		Alias string `json:"alias"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&member); err != nil {
		t.Fatal(err)
	}
	if member.Alias != "测试用户" {
		t.Fatalf("expected alias 测试用户, got %q", member.Alias)
	}
}

func TestBatchImportHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"rows":[{"name":"导入用户","phone":"13900000000","email":"import@example.com","departmentName":"后端组"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/org/members/batch-import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
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

func TestBatchInviteHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(fmt.Sprintf(`{"ids":["%s"]}`, contract.IDMember1.String()))
	req := httptest.NewRequest(http.MethodPost, "/api/org/members/batch-invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result struct {
		Sent int `json:"sent"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Sent < 0 {
		t.Fatalf("unexpected sent %d", result.Sent)
	}
}
