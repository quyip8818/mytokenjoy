package platformkey

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func SyncUpdatePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID string, targetActive *bool) error {
	if !syncdeps.Enabled(d) {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPIKeyID == nil {
		return fmt.Errorf("platform key mapping missing for %s", platformKeyID)
	}
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, d.Store.BudgetConsumed(), d.Store.Org(), d.Store.Budget(), d.Store.Keys(), d.Cfg.Clock())
	if err != nil {
		return err
	}
	key, ok := budgetCtx.FindPlatformKey(platformKeyID)
	if !ok {
		return fmt.Errorf("platform key not found")
	}
	departments, err := common.LoadDepartments(ctx, d.Store.Org().Nodes())
	if err != nil {
		return err
	}
	rules, err := common.LoadRoutingRules(ctx, d.Store.Org().Nodes(), d.Store.Models().Allowlist())
	if err != nil {
		return err
	}
	models, err := d.Store.Models().Models(ctx)
	if err != nil {
		return err
	}
	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, models)
	effectiveIDs := newapiunits.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	effectiveCallTypes := newapiunits.EffectiveCallTypes(models, effectiveIDs)
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, d.Store.Org().Nodes(), mapping.DepartmentID, d.Cfg.Clock())
	if err != nil {
		return err
	}
	remainPoint, err := pkgbudget.ComputeRemainForMapping(
		ctx, budgetCtx, d.Store.BudgetConsumed(), d.Store.Org(), d.Store.Budget(), d.Store.Company(), *mapping, open.String(),
	)
	if err != nil {
		return err
	}
	remainUnits := capRemainUnits(ctx, d, remainPoint, models, effectiveIDs)
	status := adminport.TokenStatusEnabled
	if targetActive != nil {
		if !*targetActive {
			status = adminport.TokenStatusDisabled
		}
	} else if key.Status != "active" {
		status = adminport.TokenStatusDisabled
	}
	remain := remainUnits
	enabled := len(effectiveCallTypes) > 0
	req := adminport.UpdateTokenInput{
		ID:                 *mapping.NewAPIKeyID,
		Name:               key.Name,
		Status:             &status,
		RemainQuota:        &remain,
		ModelLimitsEnabled: &enabled,
		ModelLimits:        newapiunits.FormatModelLimits(effectiveCallTypes),
		Group:              mapping.NewAPIGroup,
	}
	token, err := d.Client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	now := time.Now()
	remainQuota := token.RemainQuota
	return d.Mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, &remainQuota, now)
}

func DisablePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID string) error {
	keys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == platformKeyID {
			keys[i].Status = "disabled"
			if err := d.Store.Keys().SetPlatformKeys(ctx, keys); err != nil {
				return err
			}
			break
		}
	}
	if !syncdeps.Enabled(d) {
		return nil
	}
	mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPIKeyID == nil {
		return nil
	}
	status := adminport.TokenStatusDisabled
	zero := int64(0)
	req := adminport.UpdateTokenInput{
		ID:          *mapping.NewAPIKeyID,
		Status:      &status,
		RemainQuota: &zero,
	}
	_, err = d.Client.UpdateToken(ctx, req)
	return err
}
