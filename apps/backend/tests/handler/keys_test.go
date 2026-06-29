package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApprovalApproveHTTP(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPut, "/api/keys/approvals/apv-1/approve", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApprovalRejectHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"reason":"not needed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/keys/approvals/apv-2/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformKeyCreateHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"name":"http-key","memberId":"m-1","quota":200,"modelWhitelist":["gpt-4o"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/keys/platform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformKeyDeleteHTTP(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/keys/platform/plk-1", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestProviderKeyDeleteNotFoundHTTP(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/keys/provider/missing-provider-key", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformKeyDeleteNotFoundHTTP(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/keys/platform/missing-platform-key", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApprovalCreateHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"type":"quota","reason":"need more","requestedQuota":500,"memberId":"m-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/keys/approvals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testutil.SessionCookie("m-1"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
