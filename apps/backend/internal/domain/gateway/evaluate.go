package gateway

import (
	"errors"
	"fmt"
	"slices"
	"time"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

var ErrBudgetExhausted = errors.New("budget exhausted")

const minEstimatePoint = 0.01 * float64(common.DefaultPointsPerUnit)

func Evaluate(pc PrecheckContext, model string, opts PrecheckOpts) error {
	return EvaluateAt(pc, model, opts, time.Now())
}

func EvaluateAt(pc PrecheckContext, model string, opts PrecheckOpts, now time.Time) error {
	if !opts.SkipModelCheck && model == "" {
		return fmt.Errorf("model field is required")
	}
	if domaincompany.IsGatewayBlocked(pc.Wallet.CompanyStatus) {
		return fmt.Errorf("company not active")
	}
	if pc.Wallet.WalletRemain < minEstimatePoint {
		return fmt.Errorf("insufficient wallet points")
	}
	if pc.Routing.KeyStatus != "active" {
		return fmt.Errorf("platform key inactive")
	}
	if pc.Routing.KeyExpiresAt != nil && !pc.Routing.KeyExpiresAt.After(now) {
		return fmt.Errorf("platform key expired")
	}
	if !opts.SkipModelCheck && !opts.SkipModelAllowlist {
		if err := checkPlatformKey(pc.Routing, model); err != nil {
			return err
		}
	}
	return checkSoftBudget(pc.Budget)
}

func checkSoftBudget(budget BudgetState) error {
	if budget.SoftRemain == nil {
		return nil
	}
	if *budget.SoftRemain <= 0 {
		return ErrBudgetExhausted
	}
	return nil
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
