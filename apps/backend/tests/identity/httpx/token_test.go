package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestResolveSessionTokenCookie(t *testing.T) {
	t.Parallel()
	token := testutil.IssueSessionJWT(t, contract.DefaultCompanyID, uuid.MustParse("00000000-0000-7000-0000-000000000c01"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", httpx.SessionCookie+"="+token)
	if got := httpx.ResolveSessionToken(req); got != token {
		t.Fatalf("expected %q, got %q", token, got)
	}
}

func TestResolveSessionTokenBearer(t *testing.T) {
	t.Parallel()
	token := testutil.IssueSessionJWT(t, contract.DefaultCompanyID, uuid.MustParse("00000000-0000-7000-0000-000000000c02"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if got := httpx.ResolveSessionToken(req); got != token {
		t.Fatalf("expected %q, got %q", token, got)
	}
}

func TestResolveSessionTokenEmptyWithoutCredentials(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := httpx.ResolveSessionToken(req); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
