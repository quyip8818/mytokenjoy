package auth

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

// Mount registers the auth handler on the given router.
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Auth())
	h.RegisterRoutes(r)
}
