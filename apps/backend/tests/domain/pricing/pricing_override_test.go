package pricing_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/pricing"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func TestApplyTJPricing_GlobalPrice(t *testing.T) {
	t.Parallel()

	tokenJoyID := contract.TokenJoyCompanyID
	companyID := uuid.MustParse("00000000-aaaa-0000-0000-000000000001")

	snap := usage.EntryBuildSnapshot{
		GlobalPricing: []store.ModelPricingRow{
			{CompanyID: tokenJoyID, ModelType: "deepseek-chat", InputPrice: 2.0, OutputPrice: 8.0, EffectiveFrom: time.Now()},
		},
		CompanyPricing: nil,
		QuotaPerUnit:   500000,
	}

	entry := types.UsageLedgerEntry{
		Model:        "deepseek-chat",
		InputTokens:  10000,
		OutputTokens: 5000,
		QuotaAmount:  9999, // raw.Quota — should be overridden
	}

	result := usage.ApplyTJPricing(entry, snap, tokenJoyID)

	// cost = (10000*2.0 + 5000*8.0) / 1_000_000 = 0.06 元
	// quota = ceil(0.06 * 500000) = 30000
	wantQuota := pricing.CalcQuota(10000, 5000, 2.0, 8.0, 500000)
	if result.QuotaAmount != wantQuota {
		t.Errorf("QuotaAmount = %d, want %d", result.QuotaAmount, wantQuota)
	}
	if result.CallDetail.InputPrice != 2.0 {
		t.Errorf("InputPrice = %f, want 2.0", result.CallDetail.InputPrice)
	}
	if result.CallDetail.OutputPrice != 8.0 {
		t.Errorf("OutputPrice = %f, want 8.0", result.CallDetail.OutputPrice)
	}
	if result.CallDetail.ContractPricing {
		t.Error("expected ContractPricing=false for global price")
	}

	_ = companyID
}

func TestApplyTJPricing_ContractOverridesGlobal(t *testing.T) {
	t.Parallel()

	tokenJoyID := contract.TokenJoyCompanyID
	companyID := uuid.MustParse("00000000-aaaa-0000-0000-000000000002")

	snap := usage.EntryBuildSnapshot{
		GlobalPricing: []store.ModelPricingRow{
			{CompanyID: tokenJoyID, ModelType: "deepseek-chat", InputPrice: 2.0, OutputPrice: 8.0},
		},
		CompanyPricing: []store.ModelPricingRow{
			{CompanyID: companyID, ModelType: "deepseek-chat", InputPrice: 1.0, OutputPrice: 4.0},
		},
		QuotaPerUnit: 500000,
	}

	entry := types.UsageLedgerEntry{
		Model:        "deepseek-chat",
		InputTokens:  10000,
		OutputTokens: 5000,
		QuotaAmount:  9999,
	}

	result := usage.ApplyTJPricing(entry, snap, tokenJoyID)

	// Should use contract price: (10000*1.0 + 5000*4.0) / 1_000_000 = 0.03 元
	// quota = ceil(0.03 * 500000) = 15000
	wantQuota := pricing.CalcQuota(10000, 5000, 1.0, 4.0, 500000)
	if result.QuotaAmount != wantQuota {
		t.Errorf("QuotaAmount = %d, want %d (contract price)", result.QuotaAmount, wantQuota)
	}
	if result.CallDetail.ContractPricing != true {
		t.Error("expected ContractPricing=true for contract price")
	}
}

func TestApplyTJPricing_NoPricingKeepsRawQuota(t *testing.T) {
	t.Parallel()

	snap := usage.EntryBuildSnapshot{
		GlobalPricing:  nil,
		CompanyPricing: nil,
		QuotaPerUnit:   500000,
	}

	entry := types.UsageLedgerEntry{
		Model:        "unknown-model",
		InputTokens:  1000,
		OutputTokens: 500,
		QuotaAmount:  42,
	}

	result := usage.ApplyTJPricing(entry, snap, contract.TokenJoyCompanyID)

	if result.QuotaAmount != 42 {
		t.Errorf("QuotaAmount = %d, want 42 (unchanged raw.Quota)", result.QuotaAmount)
	}
}

func TestApplyTJPricing_ZeroQPUKeepsRawQuota(t *testing.T) {
	t.Parallel()

	snap := usage.EntryBuildSnapshot{
		GlobalPricing: []store.ModelPricingRow{
			{CompanyID: contract.TokenJoyCompanyID, ModelType: "gpt-4o", InputPrice: 15.0, OutputPrice: 60.0},
		},
		QuotaPerUnit: 0, // invalid QPU
	}

	entry := types.UsageLedgerEntry{
		Model:        "gpt-4o",
		InputTokens:  1000,
		OutputTokens: 500,
		QuotaAmount:  99,
	}

	result := usage.ApplyTJPricing(entry, snap, contract.TokenJoyCompanyID)

	if result.QuotaAmount != 99 {
		t.Errorf("QuotaAmount = %d, want 99 (unchanged due to zero QPU)", result.QuotaAmount)
	}
}
