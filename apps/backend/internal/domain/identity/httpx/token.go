package httpx

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
)

const (
	HeaderAuthzRevision = "X-Authz-Revision"
	refreshCookiePath   = "/api/auth/refresh"
)

// ponytail: cookie name 由 deployEnv + saas mode 自动派生，防止同域名不同实例互踢。
// 升级路径：生产环境用独立域名后可硬编码回常量。
var (
	SessionCookie = "tj_session"
	RefreshCookie = "tj_refresh"
)

// InitCookieNames derives cookie names from deploy env and saas mode.
// Must be called once at startup before serving requests.
func InitCookieNames(deployEnv string, saas bool) {
	envMap := map[string]string{
		"local":      "dev",
		"staging":    "stag",
		"production": "prod",
	}
	e := envMap[deployEnv]
	if e == "" {
		e = deployEnv
	}
	m := "l"
	if saas {
		m = "s"
	}
	id := e + "_" + m
	SessionCookie = "tjs_" + id
	RefreshCookie = "tjr_" + id
}

var (
	ErrNoToken      = errors.New("no session token")
	ErrInvalidToken = errors.New("invalid session token")
)

func ResolveSessionToken(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if authorization := r.Header.Get("Authorization"); strings.HasPrefix(authorization, "Bearer ") {
		if token := strings.TrimSpace(authorization[7:]); token != "" {
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

func UsedBearerAuth(r *http.Request) bool {
	if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
		return strings.TrimSpace(a[7:]) != ""
	}
	return false
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
