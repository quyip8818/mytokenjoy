package middleware

import (
	"net/http"

	"sms/backend/internal/http/response"
)

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromCtx(r.Context())
			if user == nil || !allowed[user.Role] {
				response.Error(w, http.StatusForbidden, "没有权限执行该操作")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
