package models

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.SessionHandlerBase
	service domainmodels.Service
}

func NewHandler(cfg config.Config, service domainmodels.Service, sessionSvc session.Service) *Handler {
	return &Handler{
		SessionHandlerBase: shared.NewSessionHandlerBase(cfg, sessionSvc),
		service:            service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.Cfg, r, h.SessionSvc, permission.ModelRead)
	read.Get("/", h.List)
	read.Get("/routing", h.RoutingList)
	read.Get("/routing/resolve", h.RoutingResolve)

	write := httpmiddleware.WriteRoutes(r, h.Cfg, h.SessionSvc)

	manageWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelManage))
	manageWrite.Post("/", h.Create)
	manageWrite.Put("/{id}/toggle", h.Toggle)

	whitelistWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelWhitelist))
	whitelistWrite.Put("/routing/{id}", h.RoutingUpdate)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.ListModels(r.Context())
	httputil.WriteJSON(w, http.StatusOK, models, err)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body types.CreateModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	model, err := h.service.CreateModel(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, model, err)
}

func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	var body types.ToggleModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.ToggleModel(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoutingList(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.ListRoutingRules(r.Context())
	httputil.WriteJSON(w, http.StatusOK, rules, err)
}

func (h *Handler) RoutingResolve(w http.ResponseWriter, r *http.Request) {
	deptID := r.URL.Query().Get("deptId")
	result, err := h.service.ResolveRouting(r.Context(), deptID)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) RoutingUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateRoutingRuleInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.UpdateRoutingRule(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}
