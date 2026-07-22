package approval

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainapproval "github.com/tokenjoy/backend/internal/domain/approval"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
)

type Handler struct {
	shared.ProtectedHandlerBase
	engine *domainapproval.Engine
}

func NewHandler(p httpdeps.Protected, engine *domainapproval.Engine) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		engine:               engine,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Read: anyone with submit or resolve permission can view
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.SelfApproval, permission.BudgetApprove)
	read.Get("/", h.List)
	read.Get("/{id}", h.Get)
	read.Get("/{id}/pre-check", h.PreCheck)

	write := httpmiddleware.ReadRoutes(r, h.Protected)

	// Submit: self:approval permission
	submitWrite := write.With(httpmiddleware.RequireAnyPermission(permission.SelfApproval))
	submitWrite.Post("/", h.Create)
	submitWrite.Put("/{id}/cancel", h.Cancel)

	// Resolve: budget:approve permission
	resolveWrite := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetApprove))
	resolveWrite.Put("/{id}/approve", h.Approve)
	resolveWrite.Put("/{id}/reject", h.Reject)
	resolveWrite.Put("/{id}/retry", h.Retry)
}

// --- Handlers ---

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := store.ApprovalListFilter{
		CompanyID: store.CompanyID(r.Context()),
	}
	if s := query.Get("status"); s != "" {
		status := types.ApprovalStatus(s)
		filter.Status = &status
	}
	if t := query.Get("type"); t != "" {
		approvalType := types.ApprovalType(t)
		filter.Type = &approvalType
	}
	if a := query.Get("applicantId"); a != "" {
		if id, err := uuid.Parse(a); err == nil {
			filter.ApplicantID = &id
		}
	}
	if l := query.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			filter.Limit = n
		}
	}
	if o := query.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil {
			filter.Offset = n
		}
	}

	items, total, err := h.engine.List(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	req, err := h.engine.Get(r.Context(), id)
	httputil.WriteJSON(w, http.StatusOK, req, err)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type     types.ApprovalType `json:"type"`
		Metadata json.RawMessage    `json:"metadata"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	input := domainapproval.CreateInput{
		Type:           body.Type,
		ApplicantID:    sessionCtx.Member.ID,
		ApplicantName:  sessionCtx.Member.Alias,
		DepartmentID:   sessionCtx.Member.DepartmentID,
		DepartmentName: sessionCtx.Member.DepartmentName,
		Metadata:       body.Metadata,
	}
	result, err := h.engine.Create(r.Context(), input)
	httputil.WriteJSON(w, http.StatusCreated, result, err)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	approver := domainapproval.ApproverInfo{
		ID:   sessionCtx.Member.ID,
		Name: sessionCtx.Member.Alias,
	}
	err = h.engine.Approve(r.Context(), id, approver)
	httputil.WriteVoid(w, err)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	approver := domainapproval.ApproverInfo{
		ID:   sessionCtx.Member.ID,
		Name: sessionCtx.Member.Alias,
	}
	err = h.engine.Reject(r.Context(), id, approver, body.Reason)
	httputil.WriteVoid(w, err)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	err = h.engine.Cancel(r.Context(), id, sessionCtx.Member.ID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) Retry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.engine.Retry(r.Context(), id)
	httputil.WriteVoid(w, err)
}

func (h *Handler) PreCheck(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	result, err := h.engine.PreCheck(r.Context(), id)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}
