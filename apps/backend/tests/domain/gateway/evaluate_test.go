package gateway_test

import (
	"testing"
	"time"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func precheckOpts(skipModel, skipAllowlist bool) domaingateway.PrecheckOpts {
	return domaingateway.PrecheckOpts{
		SkipModelCheck:     skipModel,
		SkipModelAllowlist: skipAllowlist,
	}
}

func TestEvaluateRejects(t *testing.T) {
	t.Parallel()
	now := time.Now()

	for _, tc := range gatewaytf.RejectionCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			pc := gatewaytf.BasePrecheckContext()
			if tc.MutatePC != nil {
				tc.MutatePC(&pc)
			}
			if err := domaingateway.EvaluateAt(pc, tc.Model, precheckOpts(false, false), now); err == nil {
				t.Fatalf("expected rejection for %s", tc.Name)
			}
		})
	}
}

func TestEvaluateAllowsFutureExpiry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	pc := gatewaytf.BasePrecheckContext()
	future := now.Add(time.Hour)
	pc.Routing.KeyExpiresAt = &future
	if err := domaingateway.EvaluateAt(pc, "gpt-4o", precheckOpts(false, false), now); err != nil {
		t.Fatalf("expected pass for future expiry, got %v", err)
	}
}

func TestEvaluateAllowsNullRemain(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.BasePrecheckContext(), "gpt-4o", precheckOpts(false, false)); err != nil {
		t.Fatalf("expected pass with NULL soft remain, got %v", err)
	}
}

func TestEvaluateAllowsPositiveRemain(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pos := int64(10)
	pc.Budget.Remain = &pos
	if err := domaingateway.Evaluate(pc, "gpt-4o", precheckOpts(false, false)); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestEvaluateModelsListingSkipsAllowlistNotBudget(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	zero := int64(0)
	pc.Budget.Remain = &zero
	if err := domaingateway.Evaluate(pc, "", precheckOpts(true, false)); err == nil {
		t.Fatal("expected budget exhausted for /v1/models path")
	}
	pos := int64(10)
	pc.Budget.Remain = &pos
	if err := domaingateway.Evaluate(pc, "", precheckOpts(true, false)); err != nil {
		t.Fatalf("expected pass without model check, got %v", err)
	}
}

func TestEvaluateAllowsModelInAllowlist(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Routing.HasAllowlist = true
	pc.Routing.AllowlistTypes = []string{"gpt-4o", "gpt-4o-mini"}
	if err := domaingateway.Evaluate(pc, "gpt-4o", precheckOpts(false, false)); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestEvaluateSkipsAllowlistForDevCatalogModel(t *testing.T) {
	t.Parallel()
	pc := gatewaytf.BasePrecheckContext()
	pc.Routing.HasAllowlist = true
	pc.Routing.AllowlistTypes = []string{"gpt-4o"}
	if err := domaingateway.Evaluate(pc, "test-model", precheckOpts(false, true)); err != nil {
		t.Fatalf("expected test catalog model to bypass allowlist, got %v", err)
	}
}

func TestEvaluatePassesWithSufficientWallet(t *testing.T) {
	t.Parallel()
	if err := domaingateway.Evaluate(gatewaytf.SufficientBudgetContext(), "gpt-4o", precheckOpts(false, false)); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}
