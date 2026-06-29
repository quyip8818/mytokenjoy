package common_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestResolveMemberIDCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.AddCookie(&http.Cookie{Name: common.SessionCookie, Value: "cookie-id"})
	if got := common.ResolveMemberID(req); got != "cookie-id" {
		t.Fatalf("expected cookie-id, got %q", got)
	}
}

func TestUsedBearerAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	if common.UsedBearerAuth(req) {
		t.Fatal("expected false without bearer header")
	}
	req.Header.Set("Authorization", "Bearer token-id")
	if !common.UsedBearerAuth(req) {
		t.Fatal("expected true with bearer header")
	}
}

func TestResolveMemberIDBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Authorization", "Bearer token-id")
	if got := common.ResolveMemberID(req); got != "token-id" {
		t.Fatalf("expected token-id, got %q", got)
	}
}

func TestResolveMemberIDEmptyWithoutCredentials(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	if got := common.ResolveMemberID(req); got != "" {
		t.Fatalf("expected empty member id, got %q", got)
	}
}

func TestResolveDemoMemberName(t *testing.T) {
	members := []types.Member{{ID: "m-1", Name: "张三"}}
	if got := common.ResolveDemoMemberName("m-1", members); got != "张三" {
		t.Fatalf("expected 张三, got %q", got)
	}
	if got := common.ResolveDemoMemberName("", members); got != "审批人" {
		t.Fatalf("expected default approver name, got %q", got)
	}
}
