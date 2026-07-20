package platformkey

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func SyncUpdatePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID, targetActive *bool) error {
	if !syncdeps.Enabled(d) {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPIKeyID == nil {
		return fmt.Errorf("platform key mapping missing for %s", platformKeyID)
	}
	keyPtr, err := d.Store.Keys().PlatformKeyByID(ctx, platformKeyID)
	if err != nil {
		return err
	}
	if keyPtr == nil {
		return fmt.Errorf("platform key not found")
	}
	key := *keyPtr
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
	_, effectiveCallTypes := resolveModelLimits(d, models, key.ModelWhitelist, deptAllowed)

	status := adminport.TokenStatusEnabled
	if targetActive != nil {
		if !*targetActive {
			status = adminport.TokenStatusDisabled
		}
	} else if key.Status != "active" {
		status = adminport.TokenStatusDisabled
	}
	enabled := len(effectiveCallTypes) > 0
	req := adminport.UpdateTokenInput{
		ID:                 *mapping.NewAPIKeyID,
		Name:               key.Name,
		Status:             &status,
		ModelLimitsEnabled: &enabled,
		ModelLimits:        newapiunits.FormatModelLimits(effectiveCallTypes),
		Group:              mapping.NewAPIGroup,
	}
	token, err := d.Client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	now := time.Now()
	return d.Mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, now)
}

func DisablePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) error {
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
	if err != nil {
		return err
	}
	if mapping == nil || mapping.NewAPIKeyID == nil {
		return nil
	}
	status := adminport.TokenStatusDisabled
	req := adminport.UpdateTokenInput{
		ID:     *mapping.NewAPIKeyID,
		Status: &status,
	}
	_, err = d.Client.UpdateToken(ctx, req)
	return err
}
