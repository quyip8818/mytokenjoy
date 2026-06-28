package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
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
