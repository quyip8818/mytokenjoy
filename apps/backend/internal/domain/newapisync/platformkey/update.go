package platformkey

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
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

	status := adminport.TokenStatusEnabled
	if targetActive != nil {
		if !*targetActive {
			status = adminport.TokenStatusDisabled
		}
	} else if key.Status != "active" {
		status = adminport.TokenStatusDisabled
	}
	req := adminport.UpdateTokenInput{
		ID:     *mapping.NewAPIKeyID,
		Name:   key.Name,
		Status: &status,
		Group:  mapping.NewAPIGroup,
	}
	token, err := d.Client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	now := time.Now()
	return d.Mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, now)
}

func DisablePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) error {
	if err := d.Store.Keys().DisablePlatformKey(ctx, platformKeyID); err != nil {
		return err
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
