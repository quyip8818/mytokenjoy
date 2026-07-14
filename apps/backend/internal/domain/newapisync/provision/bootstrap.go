package provision

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/secrets"
	"github.com/tokenjoy/backend/internal/store"
)

// Bootstrap synchronously creates NewAPI tokens for demo seed platform keys
// inserted directly by seed apply. Local-only; uses the same sync path as production create.
func Bootstrap(ctx context.Context, d syncdeps.Deps, companyID int64) error {
	if !syncdeps.Enabled(d) || !d.Cfg.AllowsDevHTTPRoutes() {
		return nil
	}
	ctx = company.WithDefaultCompany(ctx, companyID)

	if err := bootstrapDemoWalletUser(ctx, d, companyID); err != nil {
		slog.Default().Warn("bootstrap demo wallet user failed", "error", err)
	}

	unready, err := UnreadyPlatformKeyIDs(ctx, d)
	if err != nil {
		return err
	}
	if len(unready) == 0 {
		return nil
	}

	platformKeys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	budgetCtx, err := budget.LoadBudgetContext(
		ctx, d.Store.BudgetConsumed(), d.Store.Org(), d.Store.Budget(), d.Store.Keys(), d.Cfg.Clock(),
	)
	if err != nil {
		return fmt.Errorf("load budget context: %w", err)
	}

	for _, key := range platformKeys {
		if key.Status != "active" {
			continue
		}
		mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, key.ID)
		if err != nil {
			return err
		}
		if mapping != nil && mapping.SyncStatus == store.MappingSyncStatusSynced && mapping.NewAPIKeyID != nil {
			hash, ok, err := d.Store.Keys().PlatformKeyHashByID(ctx, key.ID)
			if err != nil {
				return err
			}
			if ok && hash != store.HashPlatformKey("pending:"+key.ID) {
				continue
			}
			if _, err := platformkey.TrySyncCreate(ctx, d, key.ID); err != nil {
				departmentID := platformkey.DepartmentIDForPlatformKey(key, budgetCtx)
				if departmentID == "" {
					return fmt.Errorf("repair platform key %s: %w", key.ID, err)
				}
				if _, err := platformkey.SyncPlatformKeyCreate(ctx, d, key, departmentID); err != nil {
					return fmt.Errorf("repair platform key %s: %w", key.ID, err)
				}
			}
			continue
		}
		departmentID := platformkey.DepartmentIDForPlatformKey(key, budgetCtx)
		if departmentID == "" {
			continue
		}
		if _, err := platformkey.SyncPlatformKeyCreate(ctx, d, key, departmentID); err != nil {
			return fmt.Errorf("bootstrap platform key %s: %w", key.ID, err)
		}
	}
	if err := reconcileSyncedPlatformKeyMappings(ctx, d, budgetCtx, platformKeys); err != nil {
		slog.Default().Warn("reconcile synced platform key mappings failed", "error", err)
	}
	return nil
}

func bootstrapDemoWalletUser(ctx context.Context, d syncdeps.Deps, companyID int64) error {
	if !d.Cfg.AllowsDevHTTPRoutes() || companyID != d.Cfg.LocalCompanyID {
		return nil
	}
	co, err := d.Store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return nil
	}
	if _, ok := store.ConfiguredNewAPIWalletUserID(co); ok {
		return nil
	}
	user, err := d.Client.CreateUser(ctx, adminport.CreateUserInput{
		Username:    fmt.Sprintf("company-%d", companyID),
		DisplayName: co.Name,
		Password:    secrets.RandomHex(8),
		Quota:       0,
	})
	if err != nil {
		return fmt.Errorf("create demo newapi wallet user: %w", err)
	}
	if user.ID <= 0 {
		return fmt.Errorf("create demo newapi wallet user: missing id")
	}
	if err := d.Store.Company().UpdateNewAPIWalletUserID(ctx, companyID, user.ID); err != nil {
		return err
	}
	slog.Default().Info("bootstrap demo newapi wallet user", "company_id", companyID, "newapi_user_id", user.ID)
	return nil
}

func reconcileSyncedPlatformKeyMappings(ctx context.Context, d syncdeps.Deps, budgetCtx budget.BudgetContext, platformKeys []types.PlatformKey) error {
	for _, key := range platformKeys {
		if key.Status != "active" {
			continue
		}
		mapping, err := d.Mappings.GetMappingByPlatformKeyID(ctx, key.ID)
		if err != nil {
			return err
		}
		if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
			continue
		}
		token, err := d.Client.GetToken(ctx, *mapping.NewAPIKeyID)
		if err != nil {
			slog.Default().Warn(
				"reconcile platform key mapping: token missing in newapi, recreate",
				"platform_key_id", key.ID,
				"newapi_key_id", *mapping.NewAPIKeyID,
				"error", err,
			)
			departmentID := platformkey.DepartmentIDForPlatformKey(key, budgetCtx)
			if departmentID == "" {
				continue
			}
			if _, err := platformkey.SyncPlatformKeyCreate(ctx, d, key, departmentID); err != nil {
				slog.Default().Warn("reconcile recreate platform key failed", "platform_key_id", key.ID, "error", err)
			}
			continue
		}
		expectedGroup := mapping.NewAPIGroup
		if token.Group != expectedGroup {
			if err := platformkey.SyncUpdatePlatformKey(ctx, d, key.ID, nil); err != nil {
				slog.Default().Warn(
					"reconcile platform key group mismatch update failed",
					"platform_key_id", key.ID,
					"expected_group", expectedGroup,
					"actual_group", token.Group,
					"error", err,
				)
			}
		}
	}
	return nil
}
