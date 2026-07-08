package me

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	memberAnalytics domainmemberanalytics.Service
}

func NewHandler(p httpdeps.Protected, memberAnalytics domainmemberanalytics.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		memberAnalytics:      memberAnalytics,
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
	view, err := h.memberAnalytics.GetDashboard(r.Context(), sessionCtx.Member.ID)
	httputil.WriteJSON(w, http.StatusOK, view, err)
}
