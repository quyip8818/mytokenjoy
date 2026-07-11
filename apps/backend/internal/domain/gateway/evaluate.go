package gateway

import (
	"fmt"
	"slices"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

const minEstimatePoint = 0.01 * float64(common.DefaultPointsPerUnit)

func Evaluate(pc PrecheckContext, model string, skipModelCheck bool) error {
	if !skipModelCheck && model == "" {
		return fmt.Errorf("model field is required")
	}
	if domaincompany.IsGatewayBlocked(pc.Wallet.CompanyStatus) {
		return fmt.Errorf("company not active")
	}
	if pc.Wallet.BalancePoint < minEstimatePoint {
		return fmt.Errorf("insufficient wallet balance")
	}
	if pc.Routing.KeyStatus != "active" {
		return fmt.Errorf("platform key inactive")
	}
	if skipModelCheck {
		return nil
	}
	return checkPlatformKey(pc.Routing, model)
}

func checkPlatformKey(routing RoutingState, modelName string) error {
	if modelName == "" {
		return nil
	}
	if !routing.HasAllowlist {
		return nil
	}
	if slices.Contains(routing.AllowlistTypes, modelName) {
		return nil
	}
	return fmt.Errorf("model not allowed")
}
