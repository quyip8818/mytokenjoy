package platformkey

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func upsertPendingPlatformKeyMapping(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID string) error {
	if !syncdeps.Enabled(d) {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping := store.PlatformKeyMapping{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
		MemberID:      key.MemberID,
		DepartmentID:  departmentID,
		ProjectID:     key.ProjectID,
		NewAPIGroup:   d.ChannelPolicy.ResolveNewAPIGroup(ctx, departmentID),
		SyncStatus:    store.MappingSyncStatusPending,
	}
	return d.Mappings.UpsertMapping(ctx, mapping)
}

func SyncCreatePlatformKey(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID string) error {
	if err := upsertPendingPlatformKeyMapping(ctx, d, key, departmentID); err != nil {
		return err
	}
	return d.Enqueuer.InsertNewAPISync(ctx, ports.SyncJob{
		CompanyID:     company.CompanyID(ctx),
		SubKind:       outbox.KindCreateKey,
		PlatformKeyID: key.ID,
	})
}

func TrySyncCreate(ctx context.Context, d syncdeps.Deps, platformKeyID string) (string, error) {
	if !syncdeps.Enabled(d) {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	existing, err := d.Mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil {
		return "", err
	}
	if existing != nil && existing.SyncStatus == store.MappingSyncStatusSynced && existing.NewAPIKeyID != nil {
		bearer, err := d.Client.GetTokenKey(ctx, *existing.NewAPIKeyID)
		if err != nil {
			return "", err
		}
		if bearer == "" {
			return "", fmt.Errorf("synced platform key secret unavailable")
		}
		if err := persistPlatformKeySecret(ctx, d, platformKeyID, bearer); err != nil {
			return "", err
		}
		return bearer, nil
	}
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, d.Store.BudgetConsumed(), d.Store.Org(), d.Store.Budget(), d.Store.Keys(), d.Cfg.Clock())
	if err != nil {
		return "", err
	}
	key, ok := budgetCtx.FindPlatformKey(platformKeyID)
	if !ok {
		return "", fmt.Errorf("platform key not found")
	}
	departments, err := common.LoadDepartments(ctx, d.Store.Org().Nodes())
	if err != nil {
		return "", err
	}
	rules, err := common.LoadRoutingRules(ctx, d.Store.Org().Nodes(), d.Store.Models().Allowlist())
	if err != nil {
		return "", err
	}
	models, err := d.Store.Models().Models(ctx)
	if err != nil {
		return "", err
	}
	departmentID := DepartmentIDForPlatformKey(key, budgetCtx)
	if departmentID == "" {
		return "", fmt.Errorf("department not resolved for key")
	}

	deptAllowed := common.ResolveDeptAllowedModelIDs(departmentID, departments, rules, models)
	effectiveIDs := newapiunits.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	effectiveCallTypes := newapiunits.EffectiveCallTypes(models, effectiveIDs)
	mapping := store.PlatformKeyMapping{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
		MemberID:      key.MemberID,
		DepartmentID:  departmentID,
		ProjectID:     key.ProjectID,
	}
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, d.Store.Org().Nodes(), departmentID, d.Cfg.Clock())
	if err != nil {
		return "", err
	}
	remainPoint, err := pkgbudget.ComputeRemainForMapping(
		ctx, budgetCtx, d.Store.BudgetConsumed(), d.Store.Org(), d.Store.Budget(), d.Store.Company(), mapping, open.String(),
	)
	if err != nil {
		return "", err
	}
	remainUnits, err := capRemainUnits(ctx, d, remainPoint, models, effectiveIDs)
	if err != nil {
		return "", err
	}

	group := d.ChannelPolicy.ResolveNewAPIGroup(ctx, departmentID)
	if err := d.Client.EnsureGroup(ctx, group, policy.GroupDisplayName(departmentID)); err != nil {
		return "", fmt.Errorf("ensure newapi group %s: %w", group, err)
	}

	walletUserID, err := newAPIWalletUserID(ctx, d)
	if err != nil {
		return "", err
	}
	req := adminport.CreateTokenInput{
		UserID:             walletUserID,
		Name:               TokenName(key.ID),
		RemainQuota:        remainUnits,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: len(effectiveCallTypes) > 0,
		ModelLimits:        newapiunits.FormatModelLimits(effectiveCallTypes),
		Group:              group,
		ExpiredTime:        -1,
	}
	token, err := d.Client.CreateToken(ctx, req)
	if err != nil {
		return "", err
	}
	if err := persistPlatformKeySecret(ctx, d, key.ID, token.Key); err != nil {
		return "", err
	}
	now := time.Now()
	remain := token.RemainQuota
	if err := d.Mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, &remain, now); err != nil {
		return "", err
	}
	return token.Key, nil
}

func RollbackFailedCreate(ctx context.Context, d syncdeps.Deps, platformKeyID string) {
	keys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		slog.Default().Warn("rollback failed create load keys failed", "platform_key_id", platformKeyID, "error", err)
		return
	}
	filtered := make([]types.PlatformKey, 0, len(keys))
	for _, key := range keys {
		if key.ID != platformKeyID {
			filtered = append(filtered, key)
		}
	}
	if err := d.Store.Keys().SetPlatformKeys(ctx, filtered); err != nil {
		slog.Default().Warn("rollback failed create persist failed", "platform_key_id", platformKeyID, "error", err)
	}
}

// SyncPlatformKeyCreate synchronously creates the NewAPI token and persists key_hash.
// Async retry is not enqueued; callers that need outbox durability use SyncCreatePlatformKey.
func SyncPlatformKeyCreate(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID string) (string, error) {
	if err := upsertPendingPlatformKeyMapping(ctx, d, key, departmentID); err != nil {
		return "", err
	}
	fullKey, err := TrySyncCreate(ctx, d, key.ID)
	if err != nil {
		RollbackFailedCreate(ctx, d, key.ID)
		return "", err
	}
	return fullKey, nil
}
