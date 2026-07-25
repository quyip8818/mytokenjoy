package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"sms/backend/internal/http/response"
)

type contextKey string

const userCtxKey contextKey = "user"

type AuthUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

func UserFromCtx(ctx context.Context) *AuthUser {
	u, _ := ctx.Value(userCtxKey).(*AuthUser)
	return u
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "未登录或登录已过期")
				return
			}
			tokenStr := header[7:]
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "未登录或登录已过期")
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "无效的 token")
				return
			}
			id, _ := uuid.Parse(claims["id"].(string))
			user := &AuthUser{
				ID:       id,
				Username: claims["username"].(string),
				Role:     claims["role"].(string),
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
