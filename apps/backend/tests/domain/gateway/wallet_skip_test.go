//go:build testhook

package gateway_test

import (
	"testing"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

// TestEvaluateWalletZeroRejectsWithoutSkip verifies that wallet=0 rejects
// when SkipWalletCheck is false (default behavior for SaaS/standard companies).
func TestEvaluateWalletZeroRejectsWithoutSkip(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Wallet.WalletRemainQuota = 0

	opts := domaingateway.PrecheckOpts{SkipWalletCheck: false}
	if err := domaingateway.Evaluate(pc, "gpt-4o", opts); err == nil {
		t.Fatal("expected wallet rejection when SkipWalletCheck=false and wallet=0")
	}
}

// TestEvaluateWalletZeroPassesWithSkip verifies that wallet=0 is allowed
// when SkipWalletCheck is true (selfhosted companies with non-platform channels).
func TestEvaluateWalletZeroPassesWithSkip(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Wallet.WalletRemainQuota = 0

	opts := domaingateway.PrecheckOpts{SkipWalletCheck: true}
	if err := domaingateway.Evaluate(pc, "gpt-4o", opts); err != nil {
		t.Fatalf("expected pass when SkipWalletCheck=true, got: %v", err)
	}
}

// TestEvaluateSkipWalletStillChecksCompanyStatus verifies that even with
// SkipWalletCheck=true, a suspended company is still blocked.
func TestEvaluateSkipWalletStillChecksCompanyStatus(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Wallet.WalletRemainQuota = 0
	pc.Wallet.CompanyStatus = "suspended"

	opts := domaingateway.PrecheckOpts{SkipWalletCheck: true}
	if err := domaingateway.Evaluate(pc, "gpt-4o", opts); err == nil {
		t.Fatal("expected rejection for suspended company even with SkipWalletCheck")
	}
}

// TestEvaluateSkipWalletStillChecksBudget verifies that budget exhaustion
// still blocks even when wallet check is skipped.
func TestEvaluateSkipWalletStillChecksBudget(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Wallet.WalletRemainQuota = 0
	zero := float64(0)
	pc.Budget.Remain = &zero

	opts := domaingateway.PrecheckOpts{SkipWalletCheck: true}
	if err := domaingateway.Evaluate(pc, "gpt-4o", opts); err == nil {
		t.Fatal("expected budget exhausted rejection even with SkipWalletCheck")
	}
}

// TestPrecheckSelfhostedWalletZeroPasses verifies the full PrecheckService.Run
// skips wallet check for selfhosted companies, allowing wallet=0 to pass.
func TestPrecheckSelfhostedWalletZeroPasses(t *testing.T) {
	t.Parallel()
	zeroWallet := float64(0)
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{
		Budget:             1000,
		WalletBalancePoint: &zeroWallet,
		CompanyType:        "selfhosted",
	})

	// Selfhosted + wallet=0 → should pass (SkipWalletCheck set by PrecheckService).
	if err := fx.Run("deepseek-v4-pro", false); err != nil {
		t.Fatalf("expected selfhosted wallet=0 to pass, got: %v", err)
	}
}

// TestPrecheckStandardWalletZeroRejects verifies that standard companies
// with wallet=0 are still rejected (no skip).
func TestPrecheckStandardWalletZeroRejects(t *testing.T) {
	t.Parallel()
	zeroWallet := float64(0)
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{
		Budget:             1000,
		WalletBalancePoint: &zeroWallet,
		CompanyType:        "standard",
	})

	// Standard + wallet=0 → should reject.
	if err := fx.Run("deepseek-v4-pro", false); err == nil {
		t.Fatal("expected standard wallet=0 to be rejected")
	}
}
