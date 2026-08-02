//go:build testhook

package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestApplyDiscount_NoDiscountKeepsRawQuota(t *testing.T) {
	t.Parallel()
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 1000,
	}
	result := usage.ApplyDiscount(entry, nil)
	if result.QuotaAmount != 1000 {
		t.Errorf("QuotaAmount = %d, want 1000 (no discount)", result.QuotaAmount)
	}
	if result.CallDetail.Discount != 0 {
		t.Errorf("Discount = %f, want 0 (not set)", result.CallDetail.Discount)
	}
}

func TestApplyDiscount_ExactMatch(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "gpt-4o", Discount: 0.8},
		{ModelType: "deepseek-chat", Discount: 0.9},
	}
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 1000,
	}
	result := usage.ApplyDiscount(entry, discounts)
	// 1000 * 0.8 = 800
	if result.QuotaAmount != 800 {
		t.Errorf("QuotaAmount = %d, want 800", result.QuotaAmount)
	}
	if result.CallDetail.Discount != 0.8 {
		t.Errorf("Discount = %f, want 0.8", result.CallDetail.Discount)
	}
	if !result.CallDetail.ContractPricing {
		t.Error("expected ContractPricing=true when discount applied")
	}
}

func TestApplyDiscount_WildcardFallback(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "gpt-4o", Discount: 0.7},
		{ModelType: "*", Discount: 0.9},
	}
	entry := types.UsageLedgerEntry{
		Model:       "deepseek-chat", // not in exact match
		QuotaAmount: 1000,
	}
	result := usage.ApplyDiscount(entry, discounts)
	// 1000 * 0.9 = 900 (wildcard)
	if result.QuotaAmount != 900 {
		t.Errorf("QuotaAmount = %d, want 900 (wildcard)", result.QuotaAmount)
	}
	if result.CallDetail.Discount != 0.9 {
		t.Errorf("Discount = %f, want 0.9", result.CallDetail.Discount)
	}
}

func TestApplyDiscount_ExactOverridesWildcard(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "*", Discount: 0.9},
		{ModelType: "gpt-4o", Discount: 0.5},
	}
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 1000,
	}
	result := usage.ApplyDiscount(entry, discounts)
	// 1000 * 0.5 = 500 (exact match wins)
	if result.QuotaAmount != 500 {
		t.Errorf("QuotaAmount = %d, want 500 (exact match)", result.QuotaAmount)
	}
}

func TestApplyDiscount_Markup(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "gpt-4o", Discount: 1.5},
	}
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 1000,
	}
	result := usage.ApplyDiscount(entry, discounts)
	// 1000 * 1.5 = 1500
	if result.QuotaAmount != 1500 {
		t.Errorf("QuotaAmount = %d, want 1500 (markup)", result.QuotaAmount)
	}
}

func TestApplyDiscount_CeilRounding(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "gpt-4o", Discount: 0.3},
	}
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 1, // 1 * 0.3 = 0.3, ceil = 1
	}
	result := usage.ApplyDiscount(entry, discounts)
	if result.QuotaAmount != 1 {
		t.Errorf("QuotaAmount = %d, want 1 (ceil of 0.3)", result.QuotaAmount)
	}
}

func TestApplyDiscount_DiscountOneIsNoOp(t *testing.T) {
	t.Parallel()
	discounts := []store.ModelDiscountRow{
		{ModelType: "gpt-4o", Discount: 1.0},
	}
	entry := types.UsageLedgerEntry{
		Model:       "gpt-4o",
		QuotaAmount: 999,
	}
	result := usage.ApplyDiscount(entry, discounts)
	if result.QuotaAmount != 999 {
		t.Errorf("QuotaAmount = %d, want 999 (discount=1.0 is no-op)", result.QuotaAmount)
	}
	if result.CallDetail.ContractPricing {
		t.Error("expected ContractPricing=false when discount=1.0")
	}
}
