package registertoken

import (
	"net/http"

	"github.com/google/uuid"
)

// CookieName is the shared cookie name for the register session.
const CookieName = "tokenjoy_register_session"

// SetCookie writes the register session cookie to the response.
func SetCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   600, // 10 min
	})
}

// ClearCookie removes the register session cookie.
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

// ResolveUserID reads and parses the register session cookie, returning the user ID.
func (i *Issuer) ResolveUserID(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := i.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}
