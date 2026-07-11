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
			mut:  func(pc *domaingateway.PrecheckContext) { pc.Wallet.BalancePoint = 0 },
		},
		{
			name: "zero budget",
			mut:  func(pc *domaingateway.PrecheckContext) { pc.Budget.DeptBudget = 0 },
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
			name: "key axis exhausted",
			mut: func(pc *domaingateway.PrecheckContext) {
				pc.Budget.DeptBudget = 1000
				pc.Budget.DeptConsumed = 100
				pc.Budget.KeyBudget = 200
				pc.Budget.KeyConsumed = 195
			},
		},
		{
			name: "member axis exhausted",
			mut: func(pc *domaingateway.PrecheckContext) {
				memberID := "m-1"
				pc.Budget.MemberID = &memberID
				pc.Budget.MemberFound = true
				pc.Budget.MemberCap = 1000
				pc.Budget.MemberConsumed = 999
				pc.Budget.DeptBudget = 100000
				pc.Budget.KeyBudget = 100000
			},
		},
		{
			name: "group axis exhausted",
			mut: func(pc *domaingateway.PrecheckContext) {
				groupID := "bg-1"
				pc.Budget.BudgetGroupID = &groupID
				pc.Budget.GroupBudget = 500
				pc.Budget.GroupConsumed = 495
				pc.Budget.DeptBudget = 100000
				pc.Budget.KeyBudget = 100000
			},
		},
		{
			name: "department consumed exhausted",
			mut: func(pc *domaingateway.PrecheckContext) {
				pc.Budget.DeptBudget = 100
				pc.Budget.DeptConsumed = 100
				pc.Budget.KeyBudget = 100000
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

func TestEvaluateAllowsModelsListingSkip(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.BasePrecheckContext(), "", true); err != nil {
		t.Fatalf("expected pass, got %v", err)
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

func TestEvaluatePassesWithSufficientBudget(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.SufficientBudgetContext(), "gpt-4o", false); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}
