package authz_test

import (
	"encoding/json"
	"net/http"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestRoleUpdateBumpsAuthzRevisionHeader(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	admin := testhttp.AdminCookie(t)

	sessionRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session: expected 200, got %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var before types.SessionContext
	if err := json.NewDecoder(sessionRec.Body).Decode(&before); err != nil {
		t.Fatal(err)
	}

	updateRec := testhttp.ServeAuthz(
		t, router, http.MethodPut, "/api/org/roles/role-6", admin,
		`{"name":"预算审批员","permissions":["p-6","p-12"]}`,
		nil,
	)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update role: expected 200, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	treeRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/org/departments/tree", admin, "", nil)
	if treeRec.Code != http.StatusOK {
		t.Fatalf("departments tree: expected 200, got %d body=%s", treeRec.Code, treeRec.Body.String())
	}
	revisionHeader := treeRec.Header().Get("X-Authz-Revision")
	if revisionHeader == "" {
		t.Fatal("expected X-Authz-Revision header")
	}

	sessionRec2 := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
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

func TestTransferMembersDoesNotBumpAuthzRevisionHeader(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	admin := testhttp.AdminCookie(t)

	sessionRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session: expected 200, got %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var before types.SessionContext
	if err := json.NewDecoder(sessionRec.Body).Decode(&before); err != nil {
		t.Fatal(err)
	}

	transferRec := testhttp.ServeAuthz(
		t, router, http.MethodPost, "/api/org/members/transfer",
		admin, `{"ids":["m-1"],"departmentId":"dept-4"}`, nil,
	)
	if transferRec.Code != http.StatusOK {
		t.Fatalf("transfer members: expected 200, got %d body=%s", transferRec.Code, transferRec.Body.String())
	}

	treeRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/org/departments/tree", admin, "", nil)
	if treeRec.Code != http.StatusOK {
		t.Fatalf("departments tree: expected 200, got %d body=%s", treeRec.Code, treeRec.Body.String())
	}
	revisionHeader := treeRec.Header().Get("X-Authz-Revision")
	if revisionHeader == "" {
		t.Fatal("expected X-Authz-Revision header")
	}

	sessionRec2 := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", admin, "", nil)
	if sessionRec2.Code != http.StatusOK {
		t.Fatalf("session after transfer: expected 200, got %d", sessionRec2.Code)
	}
	var after types.SessionContext
	if err := json.NewDecoder(sessionRec2.Body).Decode(&after); err != nil {
		t.Fatal(err)
	}
	if after.AuthzRevision != before.AuthzRevision {
		t.Fatalf("expected authz revision unchanged after transfer: before=%d after=%d", before.AuthzRevision, after.AuthzRevision)
	}
}
