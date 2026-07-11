package gateway

import (
	"fmt"
	"slices"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
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
	if err := checkBudgetRemain(pc.Budget); err != nil {
		return err
	}
	if skipModelCheck {
		return nil
	}
	if err := checkPlatformKey(pc.Routing, model); err != nil {
		return err
	}
	return nil
}

func checkBudgetRemain(budget BudgetState) error {
	if budget.DepartmentID == "" {
		return fmt.Errorf("department not found")
	}
	if !budget.DeptFound || budget.DeptBudget <= 0 {
		return fmt.Errorf("budget exceeded")
	}

	key := types.PlatformKey{
		ID:            budget.PlatformKeyID,
		Budget:        budget.KeyBudget,
		Used:          budget.KeyConsumed,
		MemberID:      budget.MemberID,
		BudgetGroupID: budget.BudgetGroupID,
	}

	deptAxis := &pkgbudget.DeptAxisInput{
		Budget:   budget.DeptBudget,
		Consumed: budget.DeptConsumed,
		// Precheck omits reserved_pool; ingest uses RemainForMapping with full dept axis.
	}

	var memberAxis *pkgbudget.MemberAxisInput
	if budget.MemberID != nil && budget.BudgetGroupID == nil {
		if !budget.MemberFound {
			memberAxis = &pkgbudget.MemberAxisInput{Skip: true}
		} else {
			memberAxis = &pkgbudget.MemberAxisInput{
				Cap:      budget.MemberCap,
				Consumed: budget.MemberConsumed,
			}
		}
	}

	var groups []types.BudgetGroup
	if budget.BudgetGroupID != nil {
		groups = []types.BudgetGroup{{
			ID:       *budget.BudgetGroupID,
			Budget:   budget.GroupBudget,
			Consumed: budget.GroupConsumed,
		}}
	}

	remain := pkgbudget.ComputeRemainBudget(
		key,
		nil,
		nil,
		nil,
		groups,
		budget.DepartmentID,
		memberAxis,
		deptAxis,
	)
	if remain < minEstimatePoint {
		return fmt.Errorf("budget exceeded")
	}
	return nil
}

func checkPlatformKey(routing RoutingState, modelName string) error {
	if routing.KeyStatus != "active" {
		return fmt.Errorf("platform key inactive")
	}
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
