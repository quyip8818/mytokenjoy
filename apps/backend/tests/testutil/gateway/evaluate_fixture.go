//go:build testhook

package gatewayfix

import (
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
)

// BasePrecheckContext returns a minimal passing Evaluate input for unit tests.
func BasePrecheckContext() domaingateway.PrecheckContext {
	return domaingateway.PrecheckContext{
		Wallet: domaingateway.WalletState{
			CompanyStatus: "active",
			BalancePoint:  100000,
		},
		Budget: domaingateway.BudgetState{
			DepartmentID:  "dept-3",
			DeptFound:     true,
			DeptBudget:    1000,
			DeptConsumed:  0,
			PlatformKeyID: "plk-1",
		},
		Routing: domaingateway.RoutingState{
			KeyStatus: "active",
		},
	}
}

// SufficientBudgetContext returns a context that passes wallet and multi-axis budget checks.
func SufficientBudgetContext() domaingateway.PrecheckContext {
	pc := BasePrecheckContext()
	pc.Budget.DeptBudget = testutil.DisplayPoints(1000)
	pc.Budget.KeyBudget = testutil.DisplayPoints(500)
	pc.Wallet.BalancePoint = float64(common.DefaultPointsPerUnit) * 100
	return pc
}
