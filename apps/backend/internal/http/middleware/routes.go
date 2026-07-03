package middleware

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

func SessionRoutes(r chi.Router, p httpdeps.Protected) chi.Router {
	return r.With(RequireSession(p))
}

func ReadRoutes(r chi.Router, p httpdeps.Protected, perms ...string) chi.Router {
	chain := SessionRoutes(r, p)
	if len(perms) > 0 {
		chain = chain.With(RequireAnyPermission(perms...))
	}
	return chain
}
