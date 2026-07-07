package keys

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service domainkeys.Service
}

func NewHandler(p httpdeps.Protected, service domainkeys.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	adminRead := httpmiddleware.ReadRoutes(r, h.Protected, permission.KeysRead)
	adminRead.Get("/provider", h.ProviderList)
	adminRead.Get("/platform", h.PlatformList)
	adminRead.Get("/platform/quota-summary", h.PlatformQuotaSummary)

	approvalRead := httpmiddleware.ReadRoutes(r, h.Protected, permission.KeysRead, permission.BudgetApprove)
	approvalRead.Get("/approvals", h.ApprovalsList)
	approvalRead.Get("/approvals/{id}/quota-check", h.ApprovalQuotaCheck)

	write := httpmiddleware.ReadRoutes(r, h.Protected)

	providerWrite := write.With(httpmiddleware.RequireAnyPermission(permission.KeysProvider))
	providerWrite.Post("/provider", h.ProviderCreate)
	providerWrite.Put("/provider/{id}/toggle", h.ProviderToggle)
	providerWrite.Post("/provider/{id}/rotate", h.ProviderRotate)
	providerWrite.Delete("/provider/{id}", h.ProviderDelete)

	platformWrite := write.With(httpmiddleware.RequireAnyPermission(permission.KeysAdmin))
	platformWrite.Post("/platform", h.PlatformCreate)
	platformWrite.Put("/platform/{id}", h.PlatformUpdate)
	platformWrite.Put("/platform/{id}/toggle", h.PlatformToggle)
	platformWrite.Post("/platform/{id}/rotate", h.PlatformRotate)
	platformWrite.Put("/platform/{id}/revoke", h.PlatformRevoke)
	platformWrite.Delete("/platform/{id}", h.PlatformDelete)

	approvalWrite := write.With(httpmiddleware.RequireAnyPermission(permission.SelfApproval))
	approvalWrite.Post("/approvals", h.ApprovalCreate)

	approveWrite := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetApprove))
	approveWrite.Put("/approvals/{id}/approve", h.ApprovalApprove)
	approveWrite.Put("/approvals/{id}/reject", h.ApprovalReject)
}

func (h *Handler) ProviderList(w http.ResponseWriter, r *http.Request) {
	keys, err := h.service.ListProviderKeys(r.Context())
	httputil.WriteJSON(w, http.StatusOK, keys, err)
}

func (h *Handler) ProviderCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateProviderKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.CreateProviderKey(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) ProviderToggle(w http.ResponseWriter, r *http.Request) {
	var body types.ToggleProviderKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.ToggleProviderKey(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	httputil.WriteVoid(w, err)
}

func (h *Handler) ProviderRotate(w http.ResponseWriter, r *http.Request) {
	var body types.RotateProviderKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.RotateProviderKey(r.Context(), chi.URLParam(r, "id"), body.NewKey)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) ProviderDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteProviderKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *Handler) PlatformList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	keys, err := h.service.ListPlatformKeys(r.Context(), types.PlatformKeyListFilter{
		MemberID:      query.Get("memberId"),
		BudgetGroupID: query.Get("budgetGroupId"),
		DepartmentID:  query.Get("departmentId"),
		Type:          query.Get("type"),
	})
	if err == nil {
		for i := range keys.Items {
			keys.Items[i].FullKey = nil
		}
	}
	httputil.WriteJSON(w, http.StatusOK, keys, err)
}

func (h *Handler) PlatformQuotaSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.QuotaSummary(r.Context(), r.URL.Query().Get("memberId"))
	httputil.WriteJSON(w, http.StatusOK, summary, err)
}

func (h *Handler) PlatformCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreatePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.CreatePlatformKey(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) PlatformUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdatePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.UpdatePlatformKey(r.Context(), chi.URLParam(r, "id"), body)
	if err == nil {
		key.FullKey = nil
	}
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) PlatformToggle(w http.ResponseWriter, r *http.Request) {
	var body types.TogglePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.TogglePlatformKey(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	if err == nil {
		key.FullKey = nil
	}
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) PlatformRotate(w http.ResponseWriter, r *http.Request) {
	key, err := h.service.RotatePlatformKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *Handler) PlatformRevoke(w http.ResponseWriter, r *http.Request) {
	err := h.service.RevokePlatformKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *Handler) PlatformDelete(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeletePlatformKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *Handler) ApprovalsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	approvals, err := h.service.ListApprovals(r.Context(), query.Get("tab"), query.Get("memberId"))
	httputil.WriteJSON(w, http.StatusOK, approvals, err)
}

func (h *Handler) ApprovalCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateApprovalInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	approval, err := h.service.CreateApproval(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, approval, err)
}

func (h *Handler) ApprovalQuotaCheck(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ApprovalQuotaCheck(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) ApprovalApprove(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	err := h.service.ApproveApproval(r.Context(), chi.URLParam(r, "id"), sessionCtx.Member.ID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) ApprovalReject(w http.ResponseWriter, r *http.Request) {
	var body types.RejectApprovalInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	err := h.service.RejectApproval(r.Context(), chi.URLParam(r, "id"), sessionCtx.Member.ID, body.Reason)
	httputil.WriteVoid(w, err)
}
