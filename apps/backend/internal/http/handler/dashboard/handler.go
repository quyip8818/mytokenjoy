package dashboard

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service domaindashboard.Service
}

func NewHandler(p httpdeps.Protected, service domaindashboard.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
	}
}

func (h *Handler) CostSummary(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(r.URL.Query().Get("departmentId"))
		result, err := h.service.CostSummary(ctx, params, deptID, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) DepartmentCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		params := parseCostQueryParams(r)
		result, err := h.service.DepartmentCosts(ctx, query.Get("parentId"), params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) DepartmentMemberCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(chi.URLParam(r, "deptId"))
		result, err := h.service.DepartmentMemberCosts(ctx, deptID, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) DailyCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(r.URL.Query().Get("departmentId"))
		result, err := h.service.DailyCosts(ctx, params, deptID, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) TopConsumers(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		limit := common.ParseIntParam(query.Get("limit"), 5)
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(query.Get("departmentId"))
		result, err := h.service.TopConsumers(ctx, limit, params, deptID, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) ModelUsage(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(r.URL.Query().Get("departmentId"))
		result, err := h.service.ModelUsage(ctx, params, deptID, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) DepartmentUsage(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		deptID, _ := uuid.Parse(r.URL.Query().Get("departmentId"))
		result, err := h.service.DepartmentUsage(ctx, params, deptID, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) UsageSeries(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		result, err := h.service.UsageSeriesFromQuery(
			ctx,
			query.Get("granularity"),
			query.Get("start"),
			query.Get("end"),
			query.Get("groupBy"),
			query.Get("departmentId"),
			query.Get("memberId"),
			scope,
		)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	costRead := httpmiddleware.ReadRoutes(r, h.Protected, permission.DashboardCost)
	costRead.Get("/cost/summary", h.CostSummary)
	costRead.Get("/cost/departments", h.DepartmentCosts)
	costRead.Get("/cost/departments/{deptId}/members", h.DepartmentMemberCosts)
	costRead.Get("/cost/daily", h.DailyCosts)
	costRead.Get("/cost/top", h.TopConsumers)

	usageRead := httpmiddleware.ReadRoutes(r, h.Protected, permission.DashboardUsage)
	usageRead.Get("/usage/models", h.ModelUsage)
	usageRead.Get("/usage/teams", h.DepartmentUsage)
	usageRead.Get("/usage/series", h.UsageSeries)
}

func (h *Handler) withScope(w http.ResponseWriter, r *http.Request, fn func(context.Context, domainusage.SessionScope)) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	fn(r.Context(), domainusage.SessionScope{
		MemberID:     sessionCtx.Member.ID,
		DepartmentID: sessionCtx.Member.DepartmentID,
		Permissions:  sessionCtx.Permissions,
	})
}

func parseCostQueryParams(r *http.Request) types.CostQueryParams {
	query := r.URL.Query()
	return types.CostQueryParams{
		Period:      query.Get("period"),
		StartDate:   query.Get("startDate"),
		EndDate:     query.Get("endDate"),
		Granularity: query.Get("granularity"),
	}
}
