package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/http/response"
	pkg "github.com/tokenjoy/backend/internal/pkg"
)

type DashboardHandler struct {
	service domaindashboard.Service
}

func NewDashboardHandler(service domaindashboard.Service) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) CostSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	response.JSON(w, http.StatusOK, h.service.CostSummary(period))
}

func (h *DashboardHandler) DepartmentCosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	response.JSON(w, http.StatusOK, h.service.DepartmentCosts(query.Get("parentId"), query.Get("period")))
}

func (h *DashboardHandler) DepartmentMemberCosts(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	response.JSON(w, http.StatusOK, h.service.DepartmentMemberCosts(chi.URLParam(r, "deptId"), period))
}

func (h *DashboardHandler) DailyCosts(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	response.JSON(w, http.StatusOK, h.service.DailyCosts(period))
}

func (h *DashboardHandler) TopConsumers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := pkg.ParseIntParam(query.Get("limit"), 5)
	response.JSON(w, http.StatusOK, h.service.TopConsumers(limit, query.Get("period")))
}

func (h *DashboardHandler) ModelUsage(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ModelUsage())
}

func (h *DashboardHandler) TeamUsage(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.TeamUsage())
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	r.Get("/cost/summary", h.CostSummary)
	r.Get("/cost/departments", h.DepartmentCosts)
	r.Get("/cost/departments/{deptId}/members", h.DepartmentMemberCosts)
	r.Get("/cost/daily", h.DailyCosts)
	r.Get("/cost/top", h.TopConsumers)
	r.Get("/usage/models", h.ModelUsage)
	r.Get("/usage/teams", h.TeamUsage)
}
