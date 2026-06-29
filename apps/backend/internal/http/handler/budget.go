package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
)

type BudgetHandler struct {
	sessionHandlerBase
	service domainbudget.Service
}

func NewBudgetHandler(cfg config.Config, service domainbudget.Service, sessionSvc session.Service) *BudgetHandler {
	return &BudgetHandler{
		sessionHandlerBase: newSessionHandlerBase(cfg, sessionSvc),
		service:            service,
	}
}

func (h *BudgetHandler) Tree(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetTree())
}

func (h *BudgetHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Budget       float64  `json:"budget"`
		ReservedPool *float64 `json:"reservedPool"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	node, err := h.service.UpdateNode(r.Context(), chi.URLParam(r, "id"), body.Budget, body.ReservedPool)
	httputil.WriteJSON(w, http.StatusOK, node, err)
}

func (h *BudgetHandler) MemberQuotas(w http.ResponseWriter, r *http.Request) {
	quotas, err := h.service.ListMemberQuotas(chi.URLParam(r, "deptId"))
	httputil.WriteJSON(w, http.StatusOK, quotas, err)
}

func (h *BudgetHandler) UpdateMemberQuota(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateMemberQuotaInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	quota, err := h.service.UpdateMemberQuota(r.Context(), chi.URLParam(r, "memberId"), body.PersonalQuota)
	httputil.WriteJSON(w, http.StatusOK, quota, err)
}

func (h *BudgetHandler) GroupsList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListGroups())
}

func (h *BudgetHandler) GroupCreate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	group, err := h.service.CreateGroup(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, group, err)
}

func (h *BudgetHandler) GroupUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	group, err := h.service.UpdateGroup(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, group, err)
}

func (h *BudgetHandler) GroupDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteGroup(chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *BudgetHandler) OverrunPolicyGet(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetOverrunPolicy())
}

func (h *BudgetHandler) OverrunPolicyUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.OverrunPolicyConfig
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	policy, err := h.service.UpdateOverrunPolicy(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, policy, err)
}

func (h *BudgetHandler) AlertsList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListAlerts())
}

func (h *BudgetHandler) AlertCreate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.CreateAlert(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}

func (h *BudgetHandler) AlertUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.UpdateAlert(chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}

func (h *BudgetHandler) AlertDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteAlert(chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *BudgetHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.cfg, r, h.sessionSvc, permission.BudgetRead)
	read.Get("/tree", h.Tree)
	read.Get("/departments/{deptId}/member-quotas", h.MemberQuotas)
	read.Get("/groups", h.GroupsList)
	read.Get("/overrun-policy", h.OverrunPolicyGet)
	read.Get("/alerts", h.AlertsList)

	write := httpmiddleware.WriteRoutes(r, h.cfg, h.sessionSvc)

	allocateWrite := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetAllocate))
	allocateWrite.Put("/nodes/{id}", h.UpdateNode)
	allocateWrite.Put("/members/{memberId}", h.UpdateMemberQuota)
	allocateWrite.Post("/groups", h.GroupCreate)
	allocateWrite.Put("/groups/{id}", h.GroupUpdate)
	allocateWrite.Delete("/groups/{id}", h.GroupDelete)

	policyWrite := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetPolicy))
	policyWrite.Put("/overrun-policy", h.OverrunPolicyUpdate)
	policyWrite.Post("/alerts", h.AlertCreate)
	policyWrite.Put("/alerts/{id}", h.AlertUpdate)
	policyWrite.Delete("/alerts/{id}", h.AlertDelete)
}
