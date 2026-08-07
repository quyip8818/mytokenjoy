package platform

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
)

// --- Platform Admin: Global Pricing ---

type pricingDTO struct {
	ModelType       string  `json:"modelType"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

// ListGlobalPricing returns all current global prices from NewAPI (SOT).
func (h *Handler) ListGlobalPricing(w http.ResponseWriter, r *http.Request) {
	ratios, err := h.p.AdminPort.ListModelPricing(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]pricingDTO, 0, len(ratios))
	for _, r := range ratios {
		inputPrice, outputPrice, cacheInputPrice := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio, r.CacheRatio)
		out = append(out, pricingDTO{ModelType: r.ModelName, InputPrice: inputPrice, OutputPrice: outputPrice, CacheInputPrice: cacheInputPrice})
	}
	response.JSON(w, http.StatusOK, out)
}

type setPricingInput struct {
	ModelType       string  `json:"modelType"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

// SetGlobalPricing updates global price for a model (writes to NewAPI SOT).
func (h *Handler) SetGlobalPricing(w http.ResponseWriter, r *http.Request) {
	var body setPricingInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := h.p.AdminPort.UpsertModelRatio(r.Context(), body.ModelType, body.InputPrice, body.OutputPrice, body.CacheInputPrice); err != nil {
		httputil.WriteError(w, err)
		return
	}
	h.bumpPricingVersion(r.Context())
	response.Void(w)
}

// --- Platform Admin: Discounts ---

type discountDTO struct {
	ID        string  `json:"id"`
	ModelType string  `json:"modelType"`
	Discount  float64 `json:"discount"`
}

// ListCompanyDiscounts returns current discount coefficients for a company.
func (h *Handler) ListCompanyDiscounts(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	rows, err := h.p.ModelDiscount.CurrentDiscounts(r.Context(), companyID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]discountDTO, len(rows))
	for i, row := range rows {
		out[i] = discountDTO{ID: row.ID.String(), ModelType: row.ModelType, Discount: row.Discount}
	}
	response.JSON(w, http.StatusOK, out)
}

type setDiscountInput struct {
	ModelType string  `json:"modelType"`
	Discount  float64 `json:"discount"`
	Note      string  `json:"note"`
}

// SetCompanyDiscount sets a discount coefficient for a company+model.
// If discount == 1.0, the entry is deleted (100% = no discount).
func (h *Handler) SetCompanyDiscount(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body setDiscountInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.ModelType == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "modelType required")
		return
	}
	if body.Discount <= 0 {
		httputil.WriteStatus(w, http.StatusBadRequest, "discount must be positive")
		return
	}

	if body.Discount == 1.0 {
		// ponytail: 100% means no discount — remove from DB rather than storing a no-op row
		if err := h.p.ModelDiscount.DeleteByCompanyAndModel(r.Context(), companyID, body.ModelType); err != nil {
			httputil.WriteError(w, err)
			return
		}
	} else {
		row := store.ModelDiscountRow{
			CompanyID: companyID,
			ModelType: body.ModelType,
			Discount:  body.Discount,
			Note:      body.Note,
		}
		if err := h.p.ModelDiscount.Insert(r.Context(), row); err != nil {
			httputil.WriteError(w, err)
			return
		}
	}
	_, _ = h.p.SyncVersions.Increment(r.Context(), companyID, "discounts")
	response.Void(w)
}

// BatchSetCompanyDiscounts replaces discount entries for a company atomically.
// Each entry with discount == 1.0 is treated as deletion.
func (h *Handler) BatchSetCompanyDiscounts(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body []setDiscountInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	entries := make([]domainbilling.DiscountSetEntry, 0, len(body))
	for _, item := range body {
		entries = append(entries, domainbilling.DiscountSetEntry{
			ModelType: item.ModelType,
			Discount:  item.Discount,
			Note:      item.Note,
		})
	}
	if err := h.p.BillingSvc.BatchSetDiscounts(r.Context(), companyID, entries); err != nil {
		httputil.WriteError(w, err)
		return
	}
	response.Void(w)
}
