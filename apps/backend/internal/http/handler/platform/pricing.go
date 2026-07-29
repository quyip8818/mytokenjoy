package platform

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
)

// --- Platform Admin: Global Pricing ---

type pricingDTO struct {
	ModelType     string  `json:"modelType"`
	InputPrice    float64 `json:"inputPrice"`
	OutputPrice   float64 `json:"outputPrice"`
	Note          string  `json:"note,omitempty"`
	EffectiveFrom string  `json:"effectiveFrom,omitempty"`
}

type pricingHistoryDTO struct {
	ID            string  `json:"id"`
	ModelType     string  `json:"modelType"`
	InputPrice    float64 `json:"inputPrice"`
	OutputPrice   float64 `json:"outputPrice"`
	EffectiveFrom string  `json:"effectiveFrom"`
	Note          string  `json:"note,omitempty"`
	CreatedAt     string  `json:"createdAt"`
}

func rowToDTO(row store.ModelPricingRow) pricingDTO {
	return pricingDTO{
		ModelType:     row.ModelType,
		InputPrice:    row.InputPrice,
		OutputPrice:   row.OutputPrice,
		Note:          row.Note,
		EffectiveFrom: row.EffectiveFrom.Format(time.RFC3339),
	}
}

func rowToHistoryDTO(row store.ModelPricingRow) pricingHistoryDTO {
	return pricingHistoryDTO{
		ID:            row.ID.String(),
		ModelType:     row.ModelType,
		InputPrice:    row.InputPrice,
		OutputPrice:   row.OutputPrice,
		EffectiveFrom: row.EffectiveFrom.Format(time.RFC3339),
		Note:          row.Note,
		CreatedAt:     row.CreatedAt.Format(time.RFC3339),
	}
}

// ListGlobalPricing returns all current global prices.
func (h *Handler) ListGlobalPricing(w http.ResponseWriter, r *http.Request) {
	rows, err := h.p.PricingSvc.ListGlobalPricing(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]pricingDTO, len(rows))
	for i, row := range rows {
		out[i] = rowToDTO(row)
	}
	response.JSON(w, http.StatusOK, out)
}

type setPricingInput struct {
	ModelType   string  `json:"modelType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
	Note        string  `json:"note"`
}

// SetGlobalPricing creates a new global price entry.
func (h *Handler) SetGlobalPricing(w http.ResponseWriter, r *http.Request) {
	var body setPricingInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.ModelType == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "modelType required")
		return
	}
	if body.InputPrice < 0 || body.OutputPrice < 0 {
		httputil.WriteStatus(w, http.StatusBadRequest, "prices must be non-negative")
		return
	}
	if err := h.p.PricingSvc.SetGlobalPrice(r.Context(), body.ModelType, body.InputPrice, body.OutputPrice, body.Note); err != nil {
		httputil.WriteError(w, err)
		return
	}
	response.Void(w)
}

// GlobalPriceHistory returns the price timeline for a global model.
func (h *Handler) GlobalPriceHistory(w http.ResponseWriter, r *http.Request) {
	modelType := chi.URLParam(r, "modelType")
	if modelType == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "modelType required")
		return
	}
	rows, err := h.p.PricingSvc.PriceHistory(r.Context(), h.p.Cfg.TokenJoyCompanyID, modelType)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]pricingHistoryDTO, len(rows))
	for i, row := range rows {
		out[i] = rowToHistoryDTO(row)
	}
	response.JSON(w, http.StatusOK, out)
}

// --- Platform Admin: Contract Pricing ---

// ListContractPricing returns current contract prices for a company.
func (h *Handler) ListContractPricing(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	rows, err := h.p.PricingSvc.ListContractPricing(r.Context(), companyID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]pricingDTO, len(rows))
	for i, row := range rows {
		out[i] = rowToDTO(row)
	}
	response.JSON(w, http.StatusOK, out)
}

// SetContractPricing creates a new contract price entry for a company.
func (h *Handler) SetContractPricing(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body setPricingInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.ModelType == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "modelType required")
		return
	}
	if body.InputPrice < 0 || body.OutputPrice < 0 {
		httputil.WriteStatus(w, http.StatusBadRequest, "prices must be non-negative")
		return
	}
	if err := h.p.PricingSvc.SetContractPrice(r.Context(), companyID, body.ModelType, body.InputPrice, body.OutputPrice, body.Note); err != nil {
		httputil.WriteError(w, err)
		return
	}
	response.Void(w)
}

// ContractPriceHistory returns the price timeline for a company+model.
func (h *Handler) ContractPriceHistory(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	modelType := chi.URLParam(r, "modelType")
	if modelType == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "modelType required")
		return
	}
	rows, err := h.p.PricingSvc.PriceHistory(r.Context(), companyID, modelType)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]pricingHistoryDTO, len(rows))
	for i, row := range rows {
		out[i] = rowToHistoryDTO(row)
	}
	response.JSON(w, http.StatusOK, out)
}
