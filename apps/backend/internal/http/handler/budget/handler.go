package budget

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service domainbudget.Service
}

func NewHandler(p httpdeps.Protected, service domainbudget.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
	}
}

func (h *Handler) Tree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree(r.Context())
	httputil.WriteJSON(w, http.StatusOK, tree, err)
}

func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Budget       float64  `json:"budget"`
		ReservedPool *float64 `json:"reservedPool"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	node, err := h.service.UpdateNode(r.Context(), chi.URLParam(r, "departmentId"), body.Budget, body.ReservedPool)
	httputil.WriteJSON(w, http.StatusOK, node, err)
}

func (h *Handler) MemberQuotas(w http.ResponseWriter, r *http.Request) {
	quotas, err := h.service.ListMemberQuotas(r.Context(), chi.URLParam(r, "departmentId"))
	httputil.WriteJSON(w, http.StatusOK, quotas, err)
}

func (h *Handler) UpdateMemberQuota(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateMemberQuotaInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	quota, err := h.service.UpdateMemberQuota(r.Context(), chi.URLParam(r, "memberId"), body.PersonalQuota)
	httputil.WriteJSON(w, http.StatusOK, quota, err)
}

func (h *Handler) GroupsList(w http.ResponseWriter, r *http.Request) {
	groups, err := h.service.ListGroups(r.Context())
	httputil.WriteJSON(w, http.StatusOK, groups, err)
}

func (h *Handler) GroupCreate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	group, err := h.service.CreateGroup(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, group, err)
}

func (h *Handler) GroupUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	group, err := h.service.UpdateGroup(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, group, err)
}

func (h *Handler) GroupDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteGroup(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *Handler) OverrunPolicyGet(w http.ResponseWriter, r *http.Request) {
	policy, err := h.service.GetOverrunPolicy(r.Context())
	httputil.WriteJSON(w, http.StatusOK, policy, err)
}

func (h *Handler) OverrunPolicyUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.OverrunPolicyConfig
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	policy, err := h.service.UpdateOverrunPolicy(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, policy, err)
}

func (h *Handler) AlertsList(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.ListAlerts(r.Context())
	httputil.WriteJSON(w, http.StatusOK, rules, err)
}

func (h *Handler) AlertCreate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.CreateAlert(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}

func (h *Handler) AlertUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.UpdateAlert(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}

func (h *Handler) AlertDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteAlert(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.BudgetRead)
	read.Get("/tree", h.Tree)
	read.Get("/departments/{departmentId}/member-quotas", h.MemberQuotas)
	read.Get("/groups", h.GroupsList)
	read.Get("/overrun-policy", h.OverrunPolicyGet)
	read.Get("/alerts", h.AlertsList)

	write := httpmiddleware.ReadRoutes(r, h.Protected)

	allocateWrite := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetAllocate))
	allocateWrite.Put("/departments/{departmentId}", h.UpdateNode)
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
