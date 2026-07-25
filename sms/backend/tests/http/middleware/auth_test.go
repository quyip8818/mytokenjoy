package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"sms/backend/internal/http/middleware"
)

const testSecret = "test-jwt-secret"

func issueToken(id uuid.UUID, username, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"id":       id.String(),
		"username": username,
		"role":     role,
		"exp":      exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func TestAuth_MissingHeader(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_InvalidPrefix(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	token := issueToken(uuid.Must(uuid.NewV7()), "admin", "admin", time.Now().Add(-time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	// sign with different secret
	claims := jwt.MapClaims{"id": uuid.NewString(), "username": "admin", "role": "admin", "exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrong-secret"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	userID := uuid.Must(uuid.NewV7())
	token := issueToken(userID, "admin", "admin", time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var user middleware.AuthUser
	json.NewDecoder(rec.Body).Decode(&user)
	if user.ID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, user.ID)
	}
	if user.Username != "admin" {
		t.Fatalf("expected username admin, got %s", user.Username)
	}
	if user.Role != "admin" {
		t.Fatalf("expected role admin, got %s", user.Role)
	}
}

func TestAuth_GarbageToken(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt-at-all")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUserFromCtx_Nil(t *testing.T) {
	t.Parallel()
	// empty context should return nil
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := middleware.UserFromCtx(req.Context())
	if user != nil {
		t.Fatal("expected nil user from empty context")
	}
}
