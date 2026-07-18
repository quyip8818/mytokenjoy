package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/store"
)

// TokenPairParams holds the dependencies needed to issue a session token pair.
type TokenPairParams struct {
	Secret        []byte
	SessionTTLSec int
	RefreshTTLSec int
	SecureCookie  bool
	SessionStore  store.SessionRepository
}

// IssueTokenPair creates a session record and sets both access + refresh cookies.
func IssueTokenPair(ctx context.Context, w http.ResponseWriter, r *http.Request, p TokenPairParams, companyID, memberID, userID uuid.UUID) (string, error) {
	sid := sessiontoken.NewSessionID()
	refreshRaw := sid + "." + sessiontoken.RandomHex(32)

	ttl := time.Duration(p.SessionTTLSec) * time.Second
	accessToken, err := sessiontoken.IssueAccessToken(p.Secret, ttl, companyID, memberID, userID, sid)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	refreshTTL := time.Duration(p.RefreshTTLSec) * time.Second
	sess := store.Session{
		ID:        sid,
		UserID:    userID,
		MemberID:  memberID,
		CompanyID: companyID,
		TokenHash: sessiontoken.SHA256Hex(refreshRaw),
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		CreatedAt: now,
		ExpiresAt: now.Add(refreshTTL),
	}
	if err := p.SessionStore.Create(ctx, sess); err != nil {
		return "", err
	}

	SetSessionCookie(w, accessToken, p.SecureCookie)
	SetRefreshCookie(w, refreshRaw, p.SecureCookie, p.RefreshTTLSec)
	return accessToken, nil
}
