package newapisync

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) SyncUpdatePlatformKey(ctx context.Context, platformKeyID string, targetActive *bool) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPIKeyID == nil {
		return fmt.Errorf("platform key mapping missing for %s", platformKeyID)
	}
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, l.store.BudgetConsumed(), l.store.Org(), l.store.Budget(), l.store.Keys(), l.cfg.Clock())
	if err != nil {
		return err
	}
	key, ok := budgetCtx.FindPlatformKey(platformKeyID)
	if !ok {
		return fmt.Errorf("platform key not found")
	}
	departments, err := common.LoadDepartments(ctx, l.store.Org().Nodes())
	if err != nil {
		return err
	}
	rules, err := common.LoadRoutingRules(ctx, l.store.Org().Nodes(), l.store.Models().Allowlist())
	if err != nil {
		return err
	}
	models, err := l.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, models)
	effectiveIDs := newapiunits.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	effectiveCallTypes := newapiunits.EffectiveCallTypes(models, effectiveIDs)
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, l.store.Org().Nodes(), mapping.DepartmentID, l.cfg.Clock())
	if err != nil {
		return err
	}
	remainPoint, err := pkgbudget.ComputeRemainForMapping(
		ctx, budgetCtx, l.store.BudgetConsumed(), l.store.Org(), l.store.Budget(), l.store.Company(), *mapping, open.String(),
	)
	if err != nil {
		return err
	}
	remainUnits := l.capRemainUnits(ctx, remainPoint, models, effectiveIDs)
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
	token, err := l.client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	now := time.Now()
	remainQuota := token.RemainQuota
	return l.mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, &remainQuota, now)
}

func (l *NewAPISync) DisablePlatformKey(ctx context.Context, platformKeyID string) error {
	keys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == platformKeyID {
			keys[i].Status = "disabled"
			if err := l.store.Keys().SetPlatformKeys(ctx, keys); err != nil {
				return err
			}
			break
		}
	}
	if !l.Enabled() {
		return nil
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
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
	_, err = l.client.UpdateToken(ctx, req)
	return err
}
