package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"sms/backend/internal/http/middleware"
)

const oauthTestSecret = "oauth-test-secret"

func issueSyncToken(clientID, scope string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":   clientID,
		"scope": scope,
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(oauthTestSecret))
	return signed
}

func syncOkHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestOAuthGuard_ValidToken(t *testing.T) {
	guard := middleware.OAuthGuard(oauthTestSecret, "sync:read")
	handler := guard(http.HandlerFunc(syncOkHandler))

	token := issueSyncToken("tokenjoy-sync", "sync:read", time.Now().Add(10*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/api/sync/catalog", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestOAuthGuard_NoToken(t *testing.T) {
	guard := middleware.OAuthGuard(oauthTestSecret, "sync:read")
	handler := guard(http.HandlerFunc(syncOkHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/sync/catalog", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOAuthGuard_ExpiredToken(t *testing.T) {
	guard := middleware.OAuthGuard(oauthTestSecret, "sync:read")
	handler := guard(http.HandlerFunc(syncOkHandler))

	token := issueSyncToken("tokenjoy-sync", "sync:read", time.Now().Add(-1*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/api/sync/catalog", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOAuthGuard_WrongScope(t *testing.T) {
	guard := middleware.OAuthGuard(oauthTestSecret, "sync:read")
	handler := guard(http.HandlerFunc(syncOkHandler))

	token := issueSyncToken("tokenjoy-sync", "other:scope", time.Now().Add(10*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/api/sync/catalog", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
