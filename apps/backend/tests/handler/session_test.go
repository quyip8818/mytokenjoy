package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSessionCookieReturnsMember(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", testutil.SessionCookie(t, "m-pure"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload types.SessionContext
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Member.ID != "m-pure" {
		t.Fatalf("expected m-pure, got %s", payload.Member.ID)
	}
	if !payload.ReadOnly {
		t.Fatal("expected read-only session for m-pure")
	}
}

func TestSessionInvalidTokenReturns401(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", "tokenjoy_session_member=not-a-valid-jwt")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}
