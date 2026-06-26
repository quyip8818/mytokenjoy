package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/response"
)

type BudgetHandler struct {
	service domainbudget.Service
}

func NewBudgetHandler(service domainbudget.Service) *BudgetHandler {
	return &BudgetHandler{service: service}
}

func (h *BudgetHandler) Tree(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetTree())
}

func (h *BudgetHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Budget       float64  `json:"budget"`
		ReservedPool *float64 `json:"reservedPool"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	node, err := h.service.UpdateNode(r.Context(), chi.URLParam(r, "id"), body.Budget, body.ReservedPool)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, node)
}

func (h *BudgetHandler) MemberQuotas(w http.ResponseWriter, r *http.Request) {
	quotas, err := h.service.ListMemberQuotas(chi.URLParam(r, "deptId"))
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, quotas)
}

func (h *BudgetHandler) UpdateMemberQuota(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateMemberQuotaInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	quota, err := h.service.UpdateMemberQuota(r.Context(), chi.URLParam(r, "memberId"), body.PersonalQuota)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, quota)
}

func (h *BudgetHandler) GroupsList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListGroups())
}

func (h *BudgetHandler) GroupCreate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	group, err := h.service.CreateGroup(r.Context(), body)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, group)
}

func (h *BudgetHandler) GroupUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.BudgetGroup
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	group, err := h.service.UpdateGroup(r.Context(), chi.URLParam(r, "id"), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, group)
}

func (h *BudgetHandler) GroupDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteGroup(chi.URLParam(r, "id")); err != nil {
		if writeDomainError(w, err) {
			return
		}
	}
	response.Void(w)
}

func (h *BudgetHandler) OverrunPolicyGet(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetOverrunPolicy())
}

func (h *BudgetHandler) OverrunPolicyUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.OverrunPolicyConfig
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	policy, err := h.service.UpdateOverrunPolicy(r.Context(), body)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, policy)
}

func (h *BudgetHandler) AlertsList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListAlerts())
}

func (h *BudgetHandler) AlertCreate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	rule, err := h.service.CreateAlert(r.Context(), body)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, rule)
}

func (h *BudgetHandler) AlertUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	rule, err := h.service.UpdateAlert(chi.URLParam(r, "id"), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
	}
	response.JSON(w, http.StatusOK, rule)
}

func (h *BudgetHandler) AlertDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.service.DeleteAlert(chi.URLParam(r, "id"))
	response.Void(w)
}

func (h *BudgetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/tree", h.Tree)
	r.Put("/nodes/{id}", h.UpdateNode)
	r.Get("/departments/{deptId}/member-quotas", h.MemberQuotas)
	r.Put("/members/{memberId}", h.UpdateMemberQuota)
	r.Get("/groups", h.GroupsList)
	r.Post("/groups", h.GroupCreate)
	r.Put("/groups/{id}", h.GroupUpdate)
	r.Delete("/groups/{id}", h.GroupDelete)
	r.Get("/overrun-policy", h.OverrunPolicyGet)
	r.Put("/overrun-policy", h.OverrunPolicyUpdate)
	r.Get("/alerts", h.AlertsList)
	r.Post("/alerts", h.AlertCreate)
	r.Put("/alerts/{id}", h.AlertUpdate)
	r.Delete("/alerts/{id}", h.AlertDelete)
}
