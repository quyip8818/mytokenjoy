package me

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domainmember "github.com/tokenjoy/backend/internal/domain/member"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	memberSvc domainmember.Service
}

func NewHandler(p httpdeps.Protected, memberSvc domainmember.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		memberSvc:            memberSvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.SelfKeys)
	read.Get("/dashboard", h.GetDashboard)
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	view, err := h.memberSvc.GetDashboard(r.Context(), sessionCtx.Member.ID)
	httputil.WriteJSON(w, http.StatusOK, view, err)
}
