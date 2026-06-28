package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
)

type ModelsHandler struct {
	cfg        config.Config
	service    domainmodels.Service
	sessionSvc session.Service
}

func NewModelsHandler(cfg config.Config, service domainmodels.Service, sessionSvc session.Service) *ModelsHandler {
	return &ModelsHandler{cfg: cfg, service: service, sessionSvc: sessionSvc}
}

func (h *ModelsHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.cfg, r, h.sessionSvc, permission.ModelRead)
	read.Get("/", h.List)
	read.Get("/routing", h.RoutingList)
	read.Get("/routing/resolve", h.RoutingResolve)

	write := httpmiddleware.WriteRoutes(r, h.cfg, h.sessionSvc)

	manageWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelManage))
	manageWrite.Post("/", h.Create)
	manageWrite.Put("/{id}/toggle", h.Toggle)

	whitelistWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelWhitelist))
	whitelistWrite.Put("/routing/{id}", h.RoutingUpdate)
}

func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListModels())
}

func (h *ModelsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body types.CreateModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	model, err := h.service.CreateModel(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, model, err)
}

func (h *ModelsHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	var body types.ToggleModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.ToggleModel(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	httputil.WriteVoid(w, err)
}

func (h *ModelsHandler) RoutingList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListRoutingRules())
}

func (h *ModelsHandler) RoutingResolve(w http.ResponseWriter, r *http.Request) {
	deptID := r.URL.Query().Get("deptId")
	httputil.WriteOK(w, h.service.ResolveRouting(deptID))
}

func (h *ModelsHandler) RoutingUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateRoutingRuleInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.UpdateRoutingRule(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}
