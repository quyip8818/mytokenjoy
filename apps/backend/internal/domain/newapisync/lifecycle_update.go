package newapisync

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
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
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.store.Keys(), l.cfg.Clock())
	if err != nil {
		return err
	}
	key, ok := findPlatformKey(platformKeys, platformKeyID)
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
	members, err := l.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	groups, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.cfg.Clock())
	if err != nil {
		return err
	}
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org().Nodes(), l.cfg.Clock())
	if err != nil {
		return err
	}

	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, models)
	effectiveIDs := newapi.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	effectiveCallTypes := newapi.EffectiveCallTypes(models, effectiveIDs)
	remainCNY := ComputeRemainQuota(key, tree, members, platformKeys, groups, mapping.DepartmentID)
	remainUnits := l.capRemainUnits(ctx, remainCNY, models, effectiveIDs)
	status := newapi.TokenStatusEnabled
	if targetActive != nil {
		if !*targetActive {
			status = newapi.TokenStatusDisabled
		}
	} else if key.Status != "active" {
		status = newapi.TokenStatusDisabled
	}
	remain := remainUnits
	enabled := len(effectiveCallTypes) > 0
	req := newapi.UpdateTokenRequest{
		ID:                 *mapping.NewAPIKeyID,
		Name:               key.Name,
		Status:             &status,
		RemainQuota:        &remain,
		ModelLimitsEnabled: &enabled,
		ModelLimits:        newapi.FormatModelLimits(effectiveCallTypes),
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
	status := newapi.TokenStatusDisabled
	zero := int64(0)
	req := newapi.UpdateTokenRequest{
		ID:          *mapping.NewAPIKeyID,
		Status:      &status,
		RemainQuota: &zero,
	}
	_, err = l.client.UpdateToken(ctx, req)
	return err
}
