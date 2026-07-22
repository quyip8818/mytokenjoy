package me

import (
	"github.com/go-chi/chi/v5"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"net/http"
)

type Handler struct {
	shared.ProtectedHandlerBase
	memberAnalytics domainmemberanalytics.Service
	users           store.UserRepository
	sessions        store.SessionRepository
	verifyCode      *verifycode.Service
}

func NewHandler(p httpdeps.Protected, memberAnalytics domainmemberanalytics.Service,
	users store.UserRepository, sessions store.SessionRepository, vc *verifycode.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		memberAnalytics:      memberAnalytics,
		users:                users,
		sessions:             sessions,
		verifyCode:           vc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.SelfKeys)
	read.Get("/dashboard", h.GetDashboard)

	session := httpmiddleware.SessionRoutes(r, h.Protected)
	session.Get("/profile", h.GetProfile)
	session.Post("/change-password", h.ChangePassword)
	session.Post("/change-phone", h.ChangePhone)
	session.Post("/change-email", h.ChangeEmail)
	session.Post("/revoke-sessions", h.RevokeSessions)
	session.Get("/login-activity", h.GetLoginActivity)
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	view, err := h.memberAnalytics.GetDashboard(r.Context(), sessionCtx.Member.ID)
	httputil.WriteJSON(w, http.StatusOK, view, err)
}
