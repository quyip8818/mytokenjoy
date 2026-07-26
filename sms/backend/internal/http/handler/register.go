package handler

import (
	"github.com/go-chi/chi/v5"
	"sms/backend/internal/http/deps"
	authhandler "sms/backend/internal/http/handler/auth"
	contracthandler "sms/backend/internal/http/handler/contract"
	dashhandler "sms/backend/internal/http/handler/dashboard"
	evalhandler "sms/backend/internal/http/handler/evaluation"
	modelhandler "sms/backend/internal/http/handler/model"
	newapisynchandler "sms/backend/internal/http/handler/newapisync"
	oauthhandler "sms/backend/internal/http/handler/oauth"
	orderhandler "sms/backend/internal/http/handler/order"
	supplierhandler "sms/backend/internal/http/handler/supplier"
	synchandler "sms/backend/internal/http/handler/sync"
	userhandler "sms/backend/internal/http/handler/user"
	httpmiddleware "sms/backend/internal/http/middleware"
)

type Registry struct {
	auth       *authhandler.Handler
	supplier   *supplierhandler.Handler
	model      *modelhandler.Handler
	contract   *contracthandler.Handler
	order      *orderhandler.Handler
	eval       *evalhandler.Handler
	dash       *dashhandler.Handler
	user       *userhandler.Handler
	newapisync *newapisynchandler.Handler
	oauth      *oauthhandler.Handler
	sync       *synchandler.Handler
	deps       deps.Deps
}

func NewRegistry(d deps.Deps) Registry {
	reg := Registry{
		auth:     authhandler.NewHandler(d.Auth, d.Config),
		supplier: supplierhandler.NewHandler(d.Supplier, d.Logger),
		model:    modelhandler.NewHandler(d.Model, d.Logger),
		contract: contracthandler.NewHandler(d.Contract, d.Logger),
		order:    orderhandler.NewHandler(d.Order, d.Logger),
		eval:     evalhandler.NewHandler(d.Eval, d.Logger),
		dash:     dashhandler.NewHandler(d.Dashboard),
		user:     userhandler.NewHandler(d.User, d.Logger),
		deps:     d,
	}
	if d.NewAPISync != nil {
		reg.newapisync = newapisynchandler.NewHandler(d.NewAPISync)
	}
	if d.OAuth != nil {
		reg.oauth = oauthhandler.NewHandler(d.OAuth)
	}
	if d.Sync != nil {
		reg.sync = synchandler.NewHandler(d.Sync)
	}
	return reg
}

func (reg Registry) RegisterAPIRoutes(r chi.Router) {
	r.Route("/auth", reg.auth.RegisterRoutes)

	// OAuth2 token endpoint (public, no user auth required)
	if reg.oauth != nil {
		r.Route("/oauth", reg.oauth.RegisterRoutes)
	}

	// Sync API (protected by OAuth2 Bearer token)
	if reg.sync != nil {
		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.OAuthGuard(reg.deps.Config.JWTSecret, "sync:read"))
			r.Route("/sync", reg.sync.RegisterRoutes)
		})
	}

	r.Group(func(r chi.Router) {
		r.Use(httpmiddleware.Auth(reg.deps.Config.JWTSecret))

		r.Route("/suppliers", reg.supplier.RegisterRoutes)
		r.Route("/models", reg.model.RegisterRoutes)
		r.Route("/contracts", reg.contract.RegisterRoutes)
		r.Route("/purchase-orders", reg.order.RegisterRoutes)
		r.Route("/evaluations", reg.eval.RegisterRoutes)
		r.Route("/dashboard", reg.dash.RegisterRoutes)

		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.RequireRole("admin"))
			r.Route("/users", reg.user.RegisterRoutes)
			if reg.newapisync != nil {
				r.Route("/newapi", reg.newapisync.RegisterRoutes)
			}
		})
	})
}
