package gateway

import (
	"errors"
	"fmt"
	"slices"
	"time"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
)

var ErrBudgetExhausted = errors.New("insufficient member or key quota")

// minWalletQuota is the minimum wallet_remain (in quota units) to allow a request through.
// Equivalent to a trivially small spend threshold.
const minWalletQuota int64 = 1

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
	if pc.Wallet.WalletRemain < minWalletQuota {
		return fmt.Errorf("insufficient company wallet")
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
	return checkBudgetRemain(pc.Budget)
}

func checkBudgetRemain(budget BudgetState) error {
	if budget.Remain == nil {
		return nil
	}
	if *budget.Remain <= 0 {
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
