package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
)

type KeysHandler struct {
	service domainkeys.Service
}

func NewKeysHandler(service domainkeys.Service) *KeysHandler {
	return &KeysHandler{service: service}
}

func (h *KeysHandler) RegisterRoutes(r chi.Router) {
	r.Get("/provider", h.ProviderList)
	r.Post("/provider", h.ProviderCreate)
	r.Put("/provider/{id}/toggle", h.ProviderToggle)
	r.Post("/provider/{id}/rotate", h.ProviderRotate)
	r.Delete("/provider/{id}", h.ProviderDelete)
	r.Get("/platform", h.PlatformList)
	r.Get("/platform/quota-summary", h.PlatformQuotaSummary)
	r.Post("/platform", h.PlatformCreate)
	r.Put("/platform/{id}", h.PlatformUpdate)
	r.Put("/platform/{id}/toggle", h.PlatformToggle)
	r.Post("/platform/{id}/rotate", h.PlatformRotate)
	r.Put("/platform/{id}/revoke", h.PlatformRevoke)
	r.Delete("/platform/{id}", h.PlatformDelete)
	r.Get("/approvals", h.ApprovalsList)
	r.Post("/approvals", h.ApprovalCreate)
	r.Get("/approvals/{id}/quota-check", h.ApprovalQuotaCheck)
	r.Put("/approvals/{id}/approve", h.ApprovalApprove)
	r.Put("/approvals/{id}/reject", h.ApprovalReject)
}

func (h *KeysHandler) ProviderList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListProviderKeys())
}

func (h *KeysHandler) ProviderCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateProviderKeyInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	key, err := h.service.CreateProviderKey(r.Context(), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) ProviderToggle(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&types.ToggleProviderKeyInput{})
	if err := h.service.ToggleProviderKey(r.Context(), chi.URLParam(r, "id")); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *KeysHandler) ProviderRotate(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&types.RotateProviderKeyInput{})
	key, err := h.service.RotateProviderKey(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) ProviderDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.service.DeleteProviderKey(chi.URLParam(r, "id"))
	response.Void(w)
}

func (h *KeysHandler) PlatformList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	response.JSON(w, http.StatusOK, h.service.ListPlatformKeys(query.Get("memberId"), query.Get("budgetGroupId")))
}

func (h *KeysHandler) PlatformQuotaSummary(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.QuotaSummary(r.URL.Query().Get("memberId")))
}

func (h *KeysHandler) PlatformCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreatePlatformKeyInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	key, err := h.service.CreatePlatformKey(r.Context(), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) PlatformUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdatePlatformKeyInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	key, err := h.service.UpdatePlatformKey(r.Context(), chi.URLParam(r, "id"), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) PlatformToggle(w http.ResponseWriter, r *http.Request) {
	var body types.TogglePlatformKeyInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	key, err := h.service.TogglePlatformKey(r.Context(), chi.URLParam(r, "id"), body.Enabled)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) PlatformRotate(w http.ResponseWriter, r *http.Request) {
	key, err := h.service.RotatePlatformKey(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}

func (h *KeysHandler) PlatformRevoke(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RevokePlatformKey(r.Context(), chi.URLParam(r, "id")); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *KeysHandler) PlatformDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.service.DeletePlatformKey(chi.URLParam(r, "id"))
	response.Void(w)
}

func (h *KeysHandler) ApprovalsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	response.JSON(w, http.StatusOK, h.service.ListApprovals(query.Get("tab"), query.Get("memberId")))
}

func (h *KeysHandler) ApprovalCreate(w http.ResponseWriter, r *http.Request) {
	var body types.CreateApprovalInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	approval, err := h.service.CreateApproval(r.Context(), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, approval)
}

func (h *KeysHandler) ApprovalQuotaCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ApprovalQuotaCheck(chi.URLParam(r, "id")))
}

func (h *KeysHandler) ApprovalApprove(w http.ResponseWriter, r *http.Request) {
	approverID := sessionutil.ResolveMemberID(r)
	if err := h.service.ApproveApproval(r.Context(), chi.URLParam(r, "id"), approverID); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *KeysHandler) ApprovalReject(w http.ResponseWriter, r *http.Request) {
	var body types.RejectApprovalInput
	_ = json.NewDecoder(r.Body).Decode(&body)
	approverID := sessionutil.ResolveMemberID(r)
	if err := h.service.RejectApproval(r.Context(), chi.URLParam(r, "id"), approverID, body.Reason); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}
