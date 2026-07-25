package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"sms/backend/internal/http/middleware"
)

// reqWithRole creates a request that has passed through Auth middleware with a given role.
func reqWithRole(role string) *http.Request {
	userID := uuid.Must(uuid.NewV7())
	token := issueToken(userID, "testuser", role, time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// pipe through Auth middleware to populate context
	var enriched *http.Request
	middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enriched = r
	})).ServeHTTP(httptest.NewRecorder(), req)

	return enriched
}

func TestRequireRole_AdminAllowed(t *testing.T) {
	t.Parallel()
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithRole("admin"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_ViewerDenied(t *testing.T) {
	t.Parallel()
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithRole("viewer"))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRole_MultipleRolesAllowed(t *testing.T) {
	t.Parallel()
	handler := middleware.RequireRole("admin", "editor")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithRole("editor"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_NoUserInContext(t *testing.T) {
	t.Parallel()
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// plain request without Auth middleware — no user in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
