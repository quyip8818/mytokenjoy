package dashboard

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
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
		// Empty/invalid parentId parses to uuid.Nil, which means "show root departments".
		parentID, _ := uuid.Parse(query.Get("parentId"))
		result, err := h.service.DepartmentCosts(ctx, parentID, params, scope)
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
		limit := httputil.ParseIntParam(query.Get("limit"), 5)
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

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, grants.DashboardRead)
	read.Get("/cost/summary", h.CostSummary)
	read.Get("/cost/departments", h.DepartmentCosts)
	read.Get("/cost/departments/{deptId}/members", h.DepartmentMemberCosts)
	read.Get("/cost/daily", h.DailyCosts)
	read.Get("/cost/top", h.TopConsumers)

	read.Get("/usage/models", h.ModelUsage)
	read.Get("/usage/teams", h.DepartmentUsage)
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

// Mount registers the dashboard handler on the given router under /dashboard.
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Protected(), d.DashboardSvc)
	r.Route("/dashboard", h.RegisterRoutes)
}
