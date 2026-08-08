package platform

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainplatform "github.com/tokenjoy/backend/internal/domain/platform"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/tenant"
)

// --- Catalog API ---

// CatalogVersions returns the current catalog versions for the authenticated sync client.
func (h *Handler) CatalogVersions(w http.ResponseWriter, r *http.Request) {
	companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	global, company, err := h.p.SyncVersions.GetVersions(r.Context(), companyID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	response.JSON(w, http.StatusOK, map[string]int{
		"models":     global["models"],
		"pricing":    global["pricing"],
		"currencies": global["currencies"],
		"discounts":  company["discounts"],
		"walletLots": company["wallet_lots"],
	})
}

// catalogModelDTO is the public Catalog API response format.
// ponytail: price fields removed — pricing now synced via dedicated /sync/catalog/pricing endpoint.
type catalogModelDTO struct {
	ModelID      string   `json:"modelId"`
	DisplayName  string   `json:"displayName"`
	Provider     string   `json:"provider"`
	CallType     string   `json:"callType"`
	Capabilities []string `json:"capabilities"`
	MaxContext   int      `json:"maxContext"`
}

// CatalogModels returns the full model catalog for sync clients (no pricing — use /sync/catalog/pricing).
func (h *Handler) CatalogModels(w http.ResponseWriter, r *http.Request) {
	ctx := h.globalCtx(r.Context())

	v, _ := h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "models")

	models, err := h.p.Models.Models(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Only include non-deprecated platform-managed global models.
	var active []types.ModelInfo
	for _, m := range models {
		if !m.Deprecated && m.CompanyID == h.p.Cfg.TokenJoyCompanyID && m.Source == "platform" {
			active = append(active, m)
		}
	}

	data := make([]catalogModelDTO, 0, len(active))
	for _, m := range active {
		data = append(data, catalogModelDTO{
			ModelID:      m.Type,
			DisplayName:  m.Name,
			Provider:     m.Provider,
			CallType:     primaryCapability(m.Capabilities),
			Capabilities: m.Capabilities,
			MaxContext:   m.MaxContext,
		})
	}

	response.JSON(w, http.StatusOK, map[string]any{"version": v, "data": data})
}

// --- Platform Admin: Model CRUD ---

type createModelBody struct {
	Type            string   `json:"type"`
	Name            string   `json:"name"`
	Provider        string   `json:"provider"`
	InputPrice      float64  `json:"inputPrice"`
	OutputPrice     float64  `json:"outputPrice"`
	CacheInputPrice float64  `json:"cacheInputPrice"`
	Capabilities    []string `json:"capabilities"`
	MaxContext      int      `json:"maxContext"`
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.p.PlatformSvc.ListModelsWithPricing(r.Context())
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	response.JSON(w, http.StatusOK, models)
}

func (h *Handler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var body createModelBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.Type == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "type required")
		return
	}
	if body.Provider == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "provider required")
		return
	}

	created, err := h.p.PlatformSvc.CreateModel(r.Context(), domainplatform.CreateModelInput{
		Type:            body.Type,
		Name:            body.Name,
		Provider:        body.Provider,
		InputPrice:      body.InputPrice,
		OutputPrice:     body.OutputPrice,
		CacheInputPrice: body.CacheInputPrice,
		Capabilities:    body.Capabilities,
		MaxContext:      body.MaxContext,
	})
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("create model: %w", err))
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

type updateModelBody struct {
	Name         *string  `json:"name"`
	Type         *string  `json:"type"`
	Provider     *string  `json:"provider"`
	Deprecated   *bool    `json:"deprecated"`
	Capabilities []string `json:"capabilities"`
	MaxContext   *int     `json:"maxContext"`
}

func (h *Handler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	var body updateModelBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	updated, err := h.p.PlatformSvc.UpdateModel(r.Context(), id, domainplatform.UpdateModelInput{
		Name:         body.Name,
		Type:         body.Type,
		Provider:     body.Provider,
		Deprecated:   body.Deprecated,
		Capabilities: body.Capabilities,
		MaxContext:   body.MaxContext,
	})
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("update model: %w", err))
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

// --- Platform Admin: Pricing ---

type setPricingBody struct {
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

func (h *Handler) SetModelPricing(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	var body setPricingBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	if err := h.p.PlatformSvc.SetModelPricing(r.Context(), id, domainplatform.PricingInput{
		InputPrice: body.InputPrice, OutputPrice: body.OutputPrice, CacheInputPrice: body.CacheInputPrice,
	}); err != nil {
		httputil.WriteError(w, fmt.Errorf("set pricing: %w", err))
		return
	}
	response.Void(w)
}

// --- Platform Admin: Model Channels ---

// ListModelChannels returns the NewAPI channels that serve a given model.
func (h *Handler) ListModelChannels(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	channels, err := h.p.PlatformSvc.ListModelChannels(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, channels)
}

// --- Platform Admin: Publish ---

// PublishCatalog forces a version bump (useful for manual re-sync trigger).
func (h *Handler) PublishCatalog(w http.ResponseWriter, r *http.Request) {
	newVersion, err := h.p.PlatformSvc.PublishCatalog(r.Context())
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("publish catalog: %w", err))
		return
	}
	response.JSON(w, http.StatusOK, map[string]int{"version": newVersion})
}

// --- Helpers ---

// globalCtx returns a context with TokenJoyCompanyID set, for public endpoints
// that need to query global models without a session.
func (h *Handler) globalCtx(ctx context.Context) context.Context {
	return tenant.With(ctx, tenant.Info{CompanyID: h.p.Cfg.TokenJoyCompanyID})
}

func primaryCapability(caps []string) string {
	if len(caps) > 0 {
		return caps[0]
	}
	return "chat"
}

// SyncModelsFromNewAPI pulls the model list from NewAPI /api/pricing and upserts into TokenJoy DB.
func (h *Handler) SyncModelsFromNewAPI(w http.ResponseWriter, r *http.Request) {
	count, err := h.p.PlatformSvc.SyncFromNewAPI(r.Context())
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("sync models from newapi: %w", err))
		return
	}
	response.JSON(w, http.StatusOK, map[string]int{"synced": count})
}
