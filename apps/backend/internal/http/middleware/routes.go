package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/session"
)

func SessionRoutes(r chi.Router, cfg config.Config, sess session.Service) chi.Router {
	return r.With(RequireSession(cfg, sess))
}

func ReadRoutes(r chi.Router, cfg config.Config, sess session.Service, perms ...string) chi.Router {
	chain := SessionRoutes(r, cfg, sess)
	if len(perms) > 0 {
		chain = chain.With(RequireAnyPermission(perms...))
	}
	return chain
}

func WriteRoutes(r chi.Router, cfg config.Config, sess session.Service, perms ...string) chi.Router {
	return ReadRoutes(r, cfg, sess, perms...)
}

func PublicOrReadRoutes(cfg config.Config, r chi.Router, sess session.Service, perms ...string) chi.Router {
	if cfg.IsDemoProfile() {
		return r
	}
	return ReadRoutes(r, cfg, sess, perms...)
}
