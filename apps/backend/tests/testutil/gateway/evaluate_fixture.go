//go:build testhook

package gatewayfix

import (
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

// BasePrecheckContext returns a minimal passing Evaluate input for unit tests.
func BasePrecheckContext() domaingateway.PrecheckContext {
	return domaingateway.PrecheckContext{
		Wallet: domaingateway.WalletState{
			CompanyStatus: "active",
			WalletRemain:  100000,
		},
		Routing: domaingateway.RoutingState{
			KeyStatus: "active",
		},
	}
}

// SufficientBudgetContext returns a context that passes wallet checks.
func SufficientBudgetContext() domaingateway.PrecheckContext {
	pc := BasePrecheckContext()
	pc.Wallet.WalletRemain = float64(common.DefaultPointsPerUnit) * 100
	return pc
}
