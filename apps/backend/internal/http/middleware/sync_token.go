package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/store"
)

type syncCompanyIDKey struct{}

// WithSyncCompanyID stores the authenticated sync company ID in context.
func WithSyncCompanyID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, syncCompanyIDKey{}, id)
}

// SyncCompanyIDFromContext retrieves the sync company ID from context.
func SyncCompanyIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(syncCompanyIDKey{}).(uuid.UUID)
	return id, ok
}

// RequireSyncToken verifies the per-company sync token (cst_ prefix) from the
// Authorization header. On success it injects the company ID into context.
func RequireSyncToken(companies store.CompanyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				httputil.WriteStatus(w, http.StatusUnauthorized, "missing authorization")
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")

			// Fast reject: must be a sync token (cst_ prefix)
			if !strings.HasPrefix(token, "cst_") {
				httputil.WriteStatus(w, http.StatusUnauthorized, "invalid token format")
				return
			}

			hash := sha256Hex(token)
			co, err := companies.GetBySyncTokenHash(r.Context(), hash)
			if err != nil || co == nil {
				httputil.WriteStatus(w, http.StatusForbidden, "invalid token")
				return
			}

			if co.Status != store.CompanyStatusActive {
				httputil.WriteStatus(w, http.StatusForbidden, "company inactive")
				return
			}

			ctx := WithSyncCompanyID(r.Context(), co.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
