package httpx

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

const (
	SessionCookie       = "tokenjoy_session_member"
	RefreshCookie       = "tokenjoy_refresh"
	HeaderAuthzRevision = "X-Authz-Revision"
	refreshCookiePath   = "/api/auth/refresh"
)

var (
	ErrNoToken      = errors.New("no session token")
	ErrInvalidToken = errors.New("invalid session token")
)

func ResolveSessionToken(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token != "" {
			return token
		}
	}
	cookieHeader := r.Header.Get("Cookie")
	if cookieHeader != "" {
		re := regexp.MustCompile(SessionCookie + `=([^;]+)`)
		if match := re.FindStringSubmatch(cookieHeader); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func SetSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

func UsedBearerAuth(r *http.Request) bool {
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		return false
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")) != ""
}

func ResolveMemberClaims(r *http.Request, issuer sessiontoken.Issuer) (sessiontoken.Claims, bool) {
	token := ResolveSessionToken(r)
	if token == "" {
		return sessiontoken.Claims{}, false
	}
	claims, err := issuer.Parse(token)
	if err != nil || claims.Subject == "" {
		return sessiontoken.Claims{}, false
	}
	return claims, true
}

func ParseMemberToken(r *http.Request, issuer sessiontoken.Issuer) (sessiontoken.Claims, error) {
	claims, ok := ResolveMemberClaims(r, issuer)
	if !ok {
		if ResolveSessionToken(r) == "" {
			return sessiontoken.Claims{}, ErrNoToken
		}
		return sessiontoken.Claims{}, ErrInvalidToken
	}
	if claims.CompanyID == uuid.Nil {
		return sessiontoken.Claims{}, ErrInvalidToken
	}
	return claims, nil
}

// SetRefreshCookie writes the refresh token cookie (Path restricted to refresh endpoint).
func SetRefreshCookie(w http.ResponseWriter, token string, secure bool, maxAgeSec int) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    token,
		Path:     refreshCookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
		MaxAge:   maxAgeSec,
	})
}

// ClearRefreshCookie removes the refresh token cookie.
func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    "",
		Path:     refreshCookiePath,
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
	})
}

// ResolveRefreshCookie extracts the refresh token from the request cookie.
func ResolveRefreshCookie(r *http.Request) string {
	if cookie, err := r.Cookie(RefreshCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}
