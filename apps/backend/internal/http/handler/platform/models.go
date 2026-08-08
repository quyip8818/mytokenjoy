package platform

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
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
		Deprecated:   false,
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
		_ = h.p.AdminPort.UpsertModelRatio(r.Context(), body.Type, body.InputPrice, body.OutputPrice, body.CacheInputPrice)
		h.bumpPricingVersion(r.Context())
	}

	response.JSON(w, http.StatusCreated, created)
	h.bumpModelsCatalogVersion(r.Context())
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if body.Deprecated != nil {
		model.Deprecated = *body.Deprecated
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

	model, err := h.p.Models.ModelByID(r.Context(), id)
	if err != nil || model == nil {
		httputil.WriteStatus(w, http.StatusNotFound, "Not found")
		return
	}

	var body setPricingBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	if err := h.p.AdminPort.UpsertModelRatio(r.Context(), model.Type, body.InputPrice, body.OutputPrice, body.CacheInputPrice); err != nil {
		httputil.WriteError(w, fmt.Errorf("set pricing: %w", err))
		return
	}
	h.bumpPricingVersion(r.Context())
	response.Void(w)
}

// --- Platform Admin: Model Channels ---

// channelSummaryDTO is the response item for ListModelChannels.
type channelSummaryDTO struct {
	Name     string `json:"name"`
	Group    string `json:"group"`
	Priority int    `json:"priority"`
	Weight   int    `json:"weight"`
	Status   int    `json:"status"` // 1=enabled, 2=disabled
}

// ListModelChannels returns the NewAPI channels that serve a given model.
func (h *Handler) ListModelChannels(w http.ResponseWriter, r *http.Request) {
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

	channels, err := h.p.AdminPort.ListChannels(r.Context())
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Filter channels whose models field contains this model's type.
	var result []channelSummaryDTO
	for _, ch := range channels {
		if ch.Models == "" {
			continue
		}
		for _, m := range strings.Split(ch.Models, ",") {
			if strings.TrimSpace(m) == model.Type {
				result = append(result, channelSummaryDTO{
					Name:     ch.Name,
					Group:    ch.Group,
					Priority: ch.Priority,
					Weight:   ch.Weight,
					Status:   ch.Status,
				})
				break
			}
		}
	}
	response.JSON(w, http.StatusOK, result)
}

// --- Platform Admin: Publish ---

// PublishCatalog forces a version bump (useful for manual re-sync trigger).
func (h *Handler) PublishCatalog(w http.ResponseWriter, r *http.Request) {
	newVersion, err := h.p.SyncVersions.Increment(r.Context(), store.GlobalSyncVersion, "models")
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("publish catalog: %w", err))
		return
	}
	response.JSON(w, http.StatusOK, map[string]int{"version": newVersion})
}

// --- Helpers ---

func (h *Handler) bumpModelsCatalogVersion(ctx context.Context) {
	_, _ = h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "models")
}

func (h *Handler) bumpPricingVersion(ctx context.Context) {
	_, _ = h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "pricing")
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
		entries[i] = modelcatalog.RatioEntry{ModelName: r.ModelName, ModelRatio: r.ModelRatio, CompletionRatio: r.CompletionRatio, CacheRatio: r.CacheRatio}
	}
	modelcatalog.MergePricing(models, entries)
}

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
	ctx := h.globalCtx(r.Context())
	pricingModels, err := h.p.AdminPort.ListPricingModels(ctx)
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("sync models from newapi: %w", err))
		return
	}

	infos := make([]types.ModelInfo, 0, len(pricingModels))
	for _, pm := range pricingModels {
		// NewAPI contains all models including dev mocks; only sync real models.
		if modelcatalog.IsTestCallType(pm.ModelName) {
			continue
		}
		infos = append(infos, pricingModelToModelInfo(pm))
	}

	if err := h.p.Models.SyncFromPlatform(ctx, h.p.Cfg.TokenJoyCompanyID, infos); err != nil {
		httputil.WriteError(w, fmt.Errorf("sync models from newapi: %w", err))
		return
	}
	h.bumpModelsCatalogVersion(ctx)
	response.JSON(w, http.StatusOK, map[string]int{"synced": len(infos)})
}

// pricingModelToModelInfo converts a NewAPI PricingModel to a TokenJoy ModelInfo.
func pricingModelToModelInfo(pm adminport.PricingModel) types.ModelInfo {
	modelType := pm.ModelName
	displayName := inferDisplayName(pm.ModelName)
	provider := inferProvider(pm.ModelName)
	capabilities := parseTags(pm.Tags)
	maxContext := extractMaxContext(pm.Tags)

	return types.ModelInfo{
		Provider:     provider,
		Type:         modelType,
		Name:         displayName,
		Description:  pm.Description,
		Source:       "platform",
		Deprecated:   false,
		Capabilities: capabilities,
		MaxContext:   maxContext,
	}
}

// inferDisplayName produces a human-readable name from a model_name like "deepseek-ai/DeepSeek-OCR".
func inferDisplayName(modelName string) string {
	if idx := strings.Index(modelName, "/"); idx >= 0 {
		return modelName[idx+1:]
	}
	return modelName
}

// inferProvider extracts a provider slug from the model name.
func inferProvider(modelName string) string {
	if idx := strings.Index(modelName, "/"); idx >= 0 {
		org := strings.ToLower(modelName[:idx])
		// Normalize common org names.
		switch {
		case strings.Contains(org, "deepseek"):
			return "deepseek"
		case strings.Contains(org, "moonshot"):
			return "moonshot"
		case strings.Contains(org, "zai") || strings.Contains(org, "glm"):
			return "zhipu"
		default:
			return org
		}
	}
	lower := strings.ToLower(modelName)
	switch {
	case strings.HasPrefix(lower, "deepseek"):
		return "deepseek"
	case strings.HasPrefix(lower, "gpt") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3"):
		return "openai"
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	default:
		return ""
	}
}

// parseTags splits comma-separated tags into capabilities slice.
func parseTags(tags string) []string {
	if tags == "" {
		return []string{"chat"}
	}
	parts := strings.Split(tags, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{"chat"}
	}
	return out
}

// extractMaxContext parses max context from tags like "1M", "262.1K", "8.2K".
func extractMaxContext(tags string) int {
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		multiplier := 0
		numStr := ""
		if strings.HasSuffix(tag, "M") {
			multiplier = 1_000_000
			numStr = strings.TrimSuffix(tag, "M")
		} else if strings.HasSuffix(tag, "K") {
			multiplier = 1_000
			numStr = strings.TrimSuffix(tag, "K")
		}
		if multiplier > 0 {
			if v, err := strconv.ParseFloat(numStr, 64); err == nil {
				return int(v * float64(multiplier))
			}
		}
	}
	return 128000 // default
}
