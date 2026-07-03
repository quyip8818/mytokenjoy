package httpx

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

const (
	SessionCookie         = "tokenjoy_session_member"
	PlatformSessionCookie = "tokenjoy_platform_session"
	HeaderAuthzRevision   = "X-Authz-Revision"
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

func ResolvePlatformSessionToken(r *http.Request) string {
	if cookie, err := r.Cookie(PlatformSessionCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token != "" {
			return token
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

func SetPlatformSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     PlatformSessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func ClearPlatformSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     PlatformSessionCookie,
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
	if claims.CompanyID <= 0 {
		return sessiontoken.Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func ParsePlatformToken(r *http.Request, issuer sessiontoken.Issuer) (sessiontoken.Claims, error) {
	token := ResolvePlatformSessionToken(r)
	if token == "" {
		return sessiontoken.Claims{}, ErrNoToken
	}
	claims, err := issuer.Parse(token)
	if err != nil || claims.Subject == "" {
		return sessiontoken.Claims{}, ErrInvalidToken
	}
	return claims, nil
}
