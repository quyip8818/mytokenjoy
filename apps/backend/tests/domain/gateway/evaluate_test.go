package gateway_test

import (
	"testing"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestEvaluateRejects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		mut  func(pc *domaingateway.PrecheckContext)
	}{
		{
			name: "empty model",
			mut:  func(pc *domaingateway.PrecheckContext) {},
		},
		{
			name: "blocked company",
			mut:  func(pc *domaingateway.PrecheckContext) { pc.Wallet.CompanyStatus = "suspended" },
		},
		{
			name: "insufficient wallet",
			mut:  func(pc *domaingateway.PrecheckContext) { pc.Wallet.WalletRemain = 0 },
		},
		{
			name: "inactive key",
			mut:  func(pc *domaingateway.PrecheckContext) { pc.Routing.KeyStatus = "disabled" },
		},
		{
			name: "model not in allowlist",
			mut: func(pc *domaingateway.PrecheckContext) {
				pc.Routing.HasAllowlist = true
				pc.Routing.AllowlistTypes = []string{"gpt-4o"}
			},
		},
		{
			name: "exhausted soft remain",
			mut: func(pc *domaingateway.PrecheckContext) {
				zero := 0.0
				pc.Budget.SoftRemain = &zero
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pc := gatewaytf.BasePrecheckContext()
			tc.mut(&pc)
			model := "gpt-4o"
			if tc.name == "empty model" {
				model = ""
			}
			if tc.name == "model not in allowlist" {
				model = "unknown-model"
			}
			if err := domaingateway.Evaluate(pc, model, false); err == nil {
				t.Fatalf("expected rejection for %s", tc.name)
			}
		})
	}
}

func TestEvaluateAllowsNullSoftRemain(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.BasePrecheckContext(), "gpt-4o", false); err != nil {
		t.Fatalf("expected pass with NULL soft remain, got %v", err)
	}
}

func TestEvaluateAllowsPositiveSoftRemain(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pos := 10.0
	pc.Budget.SoftRemain = &pos
	if err := domaingateway.Evaluate(pc, "gpt-4o", false); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestEvaluateModelsListingSkipsAllowlistNotBudget(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	zero := 0.0
	pc.Budget.SoftRemain = &zero
	if err := domaingateway.Evaluate(pc, "", true); err == nil {
		t.Fatal("expected budget exhausted for /v1/models path")
	}
	pos := 10.0
	pc.Budget.SoftRemain = &pos
	if err := domaingateway.Evaluate(pc, "", true); err != nil {
		t.Fatalf("expected pass without model check, got %v", err)
	}
}

func TestEvaluateAllowsModelInAllowlist(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Routing.HasAllowlist = true
	pc.Routing.AllowlistTypes = []string{"gpt-4o", "gpt-4o-mini"}
	if err := domaingateway.Evaluate(pc, "gpt-4o", false); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestEvaluatePassesWithSufficientWallet(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.SufficientBudgetContext(), "gpt-4o", false); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}
