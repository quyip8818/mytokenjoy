package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestRoleUpdateBumpsAuthzRevisionHeader(t *testing.T) {
	router := newTestRouter(t)
	admin := adminSessionCookie(t)

	sessionRec := serveAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session: expected 200, got %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var before types.SessionContext
	if err := json.NewDecoder(sessionRec.Body).Decode(&before); err != nil {
		t.Fatal(err)
	}

	updateRec := serveAuthz(
		t, router, http.MethodPut, "/api/org/roles/role-6", admin,
		`{"name":"预算审批员","permissions":["p-6","p-12"]}`,
		nil,
	)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update role: expected 200, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	treeRec := serveAuthz(t, router, http.MethodGet, "/api/org/departments/tree", admin, "", nil)
	if treeRec.Code != http.StatusOK {
		t.Fatalf("departments tree: expected 200, got %d body=%s", treeRec.Code, treeRec.Body.String())
	}
	revisionHeader := treeRec.Header().Get("X-Authz-Revision")
	if revisionHeader == "" {
		t.Fatal("expected X-Authz-Revision header")
	}

	sessionRec2 := serveAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
	if sessionRec2.Code != http.StatusOK {
		t.Fatalf("session after update: expected 200, got %d", sessionRec2.Code)
	}
	var after types.SessionContext
	if err := json.NewDecoder(sessionRec2.Body).Decode(&after); err != nil {
		t.Fatal(err)
	}
	if after.AuthzRevision <= before.AuthzRevision {
		t.Fatalf("expected authz revision to increase: before=%d after=%d", before.AuthzRevision, after.AuthzRevision)
	}
}
