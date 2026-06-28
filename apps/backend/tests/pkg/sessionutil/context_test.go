package sessionutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
)

func TestResolveMemberIDHeaderFirst(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session?memberId=query-id", nil)
	req.Header.Set("X-Demo-Member-Id", "header-id")
	if got := sessionutil.ResolveMemberID(req, true); got != "header-id" {
		t.Fatalf("expected header-id, got %q", got)
	}
}

func TestResolveMemberIDIgnoresDemoHeaderWhenDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session?memberId=query-id", nil)
	req.Header.Set("X-Demo-Member-Id", "header-id")
	if got := sessionutil.ResolveMemberID(req, false); got != "" {
		t.Fatalf("expected empty member id, got %q", got)
	}
}

func TestResolveMemberIDCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.AddCookie(&http.Cookie{Name: sessionutil.SessionCookie, Value: "cookie-id"})
	if got := sessionutil.ResolveMemberID(req, false); got != "cookie-id" {
		t.Fatalf("expected cookie-id, got %q", got)
	}
}

func TestResolveDemoMemberName(t *testing.T) {
	members := []types.Member{{ID: "m-1", Name: "张三"}}
	if got := sessionutil.ResolveDemoMemberName("m-1", members); got != "张三" {
		t.Fatalf("expected 张三, got %q", got)
	}
	if got := sessionutil.ResolveDemoMemberName("", members); got != "审批人" {
		t.Fatalf("expected default approver name, got %q", got)
	}
}
