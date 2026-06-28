package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
)

type KeysHandler struct {
	cfg        config.Config
	service    domainkeys.Service
	sessionSvc session.Service
}

func NewKeysHandler(cfg config.Config, service domainkeys.Service, sessionSvc session.Service) *KeysHandler {
	return &KeysHandler{cfg: cfg, service: service, sessionSvc: sessionSvc}
}

func (h *KeysHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.cfg, r, h.sessionSvc, permission.KeysRead)
	read.Get("/provider", h.ProviderList)
	read.Get("/platform", h.PlatformList)
	read.Get("/platform/quota-summary", h.PlatformQuotaSummary)
	read.Get("/approvals", h.ApprovalsList)
	read.Get("/approvals/{id}/quota-check", h.ApprovalQuotaCheck)

	write := httpmiddleware.WriteRoutes(r, h.cfg, h.sessionSvc)

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

func (h *KeysHandler) ProviderList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListProviderKeys())
}

func (h *KeysHandler) ProviderCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateProviderKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.CreateProviderKey(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) ProviderToggle(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&types.ToggleProviderKeyInput{})
	err := h.service.ToggleProviderKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *KeysHandler) ProviderRotate(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&types.RotateProviderKeyInput{})
	key, err := h.service.RotateProviderKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) ProviderDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.service.DeleteProviderKey(chi.URLParam(r, "id"))
	httputil.WriteVoid(w, nil)
}

func (h *KeysHandler) PlatformList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	httputil.WriteOK(w, h.service.ListPlatformKeys(query.Get("memberId"), query.Get("budgetGroupId")))
}

func (h *KeysHandler) PlatformQuotaSummary(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.QuotaSummary(r.URL.Query().Get("memberId")))
}

func (h *KeysHandler) PlatformCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreatePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.CreatePlatformKey(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) PlatformUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdatePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.UpdatePlatformKey(r.Context(), chi.URLParam(r, "id"), body)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) PlatformToggle(w http.ResponseWriter, r *http.Request) {
	var body types.TogglePlatformKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.service.TogglePlatformKey(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) PlatformRotate(w http.ResponseWriter, r *http.Request) {
	key, err := h.service.RotatePlatformKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteJSON(w, http.StatusOK, key, err)
}

func (h *KeysHandler) PlatformRevoke(w http.ResponseWriter, r *http.Request) {
	err := h.service.RevokePlatformKey(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}

func (h *KeysHandler) PlatformDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.service.DeletePlatformKey(chi.URLParam(r, "id"))
	httputil.WriteVoid(w, nil)
}

func (h *KeysHandler) ApprovalsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	httputil.WriteOK(w, h.service.ListApprovals(query.Get("tab"), query.Get("memberId")))
}

func (h *KeysHandler) ApprovalCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateApprovalInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	approval, err := h.service.CreateApproval(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, approval, err)
}

func (h *KeysHandler) ApprovalQuotaCheck(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ApprovalQuotaCheck(chi.URLParam(r, "id")))
}

func (h *KeysHandler) ApprovalApprove(w http.ResponseWriter, r *http.Request) {
	approverID := sessionutil.ResolveMemberID(r, h.cfg.IsDemoProfile())
	err := h.service.ApproveApproval(r.Context(), chi.URLParam(r, "id"), approverID)
	httputil.WriteVoid(w, err)
}

func (h *KeysHandler) ApprovalReject(w http.ResponseWriter, r *http.Request) {
	var body types.RejectApprovalInput
	_ = json.NewDecoder(r.Body).Decode(&body)
	approverID := sessionutil.ResolveMemberID(r, h.cfg.IsDemoProfile())
	err := h.service.RejectApproval(r.Context(), chi.URLParam(r, "id"), approverID, body.Reason)
	httputil.WriteVoid(w, err)
}
