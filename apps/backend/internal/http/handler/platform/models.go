package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

// --- Catalog API (public, no auth) ---

const catalogModelsVersionKey = "catalog.models_version"
const catalogPricingVersionKey = "catalog.pricing_version"
const catalogDiscountsVersionKey = "catalog.discounts_version"
const catalogCurrenciesVersionKey = "catalog.currencies_version"
const catalogWalletLotsVersionKey = "catalog.wallet_lots_version"

// CatalogVersions returns the current catalog version for sync clients.
func (h *Handler) CatalogVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	modelsV, err := h.p.SystemSettings.Get(ctx, catalogModelsVersionKey)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	pricingV, _ := h.p.SystemSettings.Get(ctx, catalogPricingVersionKey)
	discountsV, _ := h.p.SystemSettings.Get(ctx, catalogDiscountsVersionKey)
	currenciesV, _ := h.p.SystemSettings.Get(ctx, catalogCurrenciesVersionKey)
	walletLotsV, _ := h.p.SystemSettings.Get(ctx, catalogWalletLotsVersionKey)
	mv, _ := strconv.Atoi(modelsV)     // empty → 0
	pv, _ := strconv.Atoi(pricingV)    // empty → 0
	dv, _ := strconv.Atoi(discountsV)  // empty → 0
	cv, _ := strconv.Atoi(currenciesV) // empty → 0
	wv, _ := strconv.Atoi(walletLotsV) // empty → 0
	response.JSON(w, http.StatusOK, map[string]int{"models": mv, "pricing": pv, "discounts": dv, "currencies": cv, "walletLots": wv})
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

	version, _ := h.p.SystemSettings.Get(ctx, catalogModelsVersionKey)
	v, _ := strconv.Atoi(version) // empty → 0

	models, err := h.p.Models.Models(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Only include active platform-managed global models.
	var active []types.ModelInfo
	for _, m := range models {
		if m.Active && m.CompanyID == h.p.Cfg.TokenJoyCompanyID && m.Source == "platform" {
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
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	InputPrice   float64  `json:"inputPrice"`
	OutputPrice  float64  `json:"outputPrice"`
	Capabilities []string `json:"capabilities"`
	MaxContext   int      `json:"maxContext"`
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // CompanyResolve already set TokenJoyCompanyID
	models, err := h.p.Models.Models(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Filter to global platform models only.
	var global []types.ModelInfo
	for _, m := range models {
		if m.CompanyID == h.p.Cfg.TokenJoyCompanyID {
			global = append(global, m)
		}
	}

	// Merge prices from NewAPI (SOT).
	h.mergePricing(r.Context(), global)
	response.JSON(w, http.StatusOK, global)
}

func (h *Handler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var body createModelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
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

	capabilities := body.Capabilities
	if len(capabilities) == 0 {
		capabilities = []string{"chat"}
	}
	maxContext := body.MaxContext
	if maxContext == 0 {
		maxContext = 1000000
	}

	model := types.ModelInfo{
		Provider:     body.Provider,
		Type:         body.Type,
		Name:         body.Name,
		Source:       "platform",
		Active:       true,
		Capabilities: capabilities,
		MaxContext:   maxContext,
	}
	if model.Name == "" {
		model.Name = model.Type
	}

	created, err := h.p.Models.InsertModel(r.Context(), model)
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("create model: %w", err))
		return
	}

	// Push price to NewAPI (SOT).
	if body.InputPrice > 0 || body.OutputPrice > 0 {
		_ = h.p.AdminPort.UpsertModelRatio(r.Context(), body.Type, body.InputPrice, body.OutputPrice)
		h.bumpPricingVersion(r.Context())
	}

	response.JSON(w, http.StatusCreated, created)
	h.bumpModelsCatalogVersion(r.Context())
}

type updateModelBody struct {
	Name         *string  `json:"name"`
	Type         *string  `json:"type"`
	Provider     *string  `json:"provider"`
	Active       *bool    `json:"active"`
	Capabilities []string `json:"capabilities"`
	MaxContext   *int     `json:"maxContext"`
}

func (h *Handler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	model, err := h.p.Models.ModelByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	if model == nil {
		httputil.WriteStatus(w, http.StatusNotFound, "Not found")
		return
	}

	var body updateModelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	if body.Name != nil {
		model.Name = *body.Name
	}
	if body.Type != nil {
		model.Type = *body.Type
	}
	if body.Provider != nil {
		model.Provider = *body.Provider
	}
	if body.Active != nil {
		model.Active = *body.Active
	}
	if body.Capabilities != nil {
		model.Capabilities = body.Capabilities
	}
	if body.MaxContext != nil {
		model.MaxContext = *body.MaxContext
	}

	if err := h.p.Models.UpdateModel(r.Context(), *model); err != nil {
		httputil.WriteError(w, fmt.Errorf("update model: %w", err))
		return
	}
	h.bumpModelsCatalogVersion(r.Context())
	response.JSON(w, http.StatusOK, model)
}

func (h *Handler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if err := h.p.Models.DeleteModel(r.Context(), id); err != nil {
		httputil.WriteError(w, fmt.Errorf("delete model: %w", err))
		return
	}
	h.bumpModelsCatalogVersion(r.Context())
	response.Void(w)
}

// --- Platform Admin: Pricing ---

type setPricingBody struct {
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

func (h *Handler) SetModelPricing(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	model, err := h.p.Models.ModelByID(r.Context(), id)
	if err != nil || model == nil {
		httputil.WriteStatus(w, http.StatusNotFound, "Not found")
		return
	}

	var body setPricingBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}

	if err := h.p.AdminPort.UpsertModelRatio(r.Context(), model.Type, body.InputPrice, body.OutputPrice); err != nil {
		httputil.WriteError(w, fmt.Errorf("set pricing: %w", err))
		return
	}
	h.bumpPricingVersion(r.Context())
	response.Void(w)
}

// --- Platform Admin: Publish ---

// PublishCatalog forces a version bump (useful for manual re-sync trigger).
// Under normal operation, version is auto-bumped by Create/Update/Delete.
func (h *Handler) PublishCatalog(w http.ResponseWriter, r *http.Request) {
	newVersion, err := h.p.SystemSettings.Increment(r.Context(), catalogModelsVersionKey)
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("publish catalog: %w", err))
		return
	}
	response.JSON(w, http.StatusOK, map[string]int{"version": newVersion})
}

// --- Helpers ---

func (h *Handler) bumpModelsCatalogVersion(ctx context.Context) {
	_, _ = h.p.SystemSettings.Increment(ctx, catalogModelsVersionKey)
}

func (h *Handler) bumpPricingVersion(ctx context.Context) {
	_, _ = h.p.SystemSettings.Increment(ctx, catalogPricingVersionKey)
}

// mergePricing enriches models with prices from NewAPI (SOT).
// ponytail: best-effort — failure leaves prices at 0 (non-fatal).
func (h *Handler) mergePricing(ctx context.Context, models []types.ModelInfo) {
	ratios, err := h.p.AdminPort.ListModelPricing(ctx)
	if err != nil || len(ratios) == 0 {
		return
	}
	entries := make([]modelcatalog.RatioEntry, len(ratios))
	for i, r := range ratios {
		entries[i] = modelcatalog.RatioEntry{ModelName: r.ModelName, ModelRatio: r.ModelRatio, CompletionRatio: r.CompletionRatio}
	}
	modelcatalog.MergePricing(models, entries)
}

// globalCtx returns a context with TokenJoyCompanyID set, for public endpoints
// that need to query global models without a session.
func (h *Handler) globalCtx(ctx context.Context) context.Context {
	return ctxcompany.With(ctx, ctxcompany.Info{CompanyID: h.p.Cfg.TokenJoyCompanyID})
}

func primaryCapability(caps []string) string {
	if len(caps) > 0 {
		return caps[0]
	}
	return "chat"
}
