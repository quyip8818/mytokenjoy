//go:build testhook

package gatewayfix

import (
	"net/http"
	"time"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/tests/testutil"
)

// RejectionCase is the SSOT for gateway rejection scenarios across evaluate, precheck, and handler tests.
type RejectionCase struct {
	Name     string
	MutatePC func(*domaingateway.PrecheckContext)
	Scenario GatewayScenarioOpts
	Model    string
	Precheck bool
	WantHTTP int // 0 = not tested at handler layer
}

// RejectionCases returns the shared gateway rejection scenarios.
func RejectionCases() []RejectionCase {
	zeroWallet := testutil.Float64Ptr(0)
	return []RejectionCase{
		{
			Name:  "empty model",
			Model: "",
		},
		{
			Name: "blocked company",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				pc.Wallet.CompanyStatus = "suspended"
			},
			Scenario: GatewayScenarioOpts{Budget: 1000, CompanyStatus: "suspended"},
			Model:    "deepseek-v4-pro",
			Precheck: true,
			WantHTTP: http.StatusForbidden,
		},
		{
			Name: "insufficient wallet",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				pc.Wallet.WalletRemainQuota = 0
			},
			Scenario: GatewayScenarioOpts{
				Budget:             1000,
				WalletBalancePoint: zeroWallet,
			},
			Model:    "deepseek-v4-pro",
			Precheck: true,
			WantHTTP: http.StatusForbidden,
		},
		{
			Name: "inactive key",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				pc.Routing.KeyStatus = "disabled"
			},
			Model: "gpt-4o",
		},
		{
			Name: "model not in allowlist",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				pc.Routing.HasAllowlist = true
				pc.Routing.AllowlistTypes = []string{"gpt-4o"}
			},
			Scenario: GatewayScenarioOpts{Budget: 1000},
			Model:    "unknown-model",
			Precheck: true,
			WantHTTP: http.StatusForbidden,
		},
		{
			Name: "exhausted combined key remain",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				zero := int64(0)
				pc.Budget.Remain = &zero
			},
			Model: "gpt-4o",
		},
		{
			Name: "expired key",
			MutatePC: func(pc *domaingateway.PrecheckContext) {
				expired := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
				pc.Routing.KeyExpiresAt = &expired
			},
			Model: "gpt-4o",
		},
	}
}
