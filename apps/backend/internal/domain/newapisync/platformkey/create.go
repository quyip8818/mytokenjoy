package platformkey

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func upsertPendingPlatformKeyMapping(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID uuid.UUID) error {
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

func SyncCreatePlatformKey(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID uuid.UUID) error {
	if err := upsertPendingPlatformKeyMapping(ctx, d, key, departmentID); err != nil {
		return err
	}
	return d.Enqueuer.InsertNewAPISync(ctx, ports.SyncJob{
		CompanyID:     company.CompanyID(ctx),
		SubKind:       outbox.KindCreateKey,
		PlatformKeyID: key.ID,
	})
}

func TrySyncCreate(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) (string, error) {
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
	keyPtr, err := d.Store.Keys().PlatformKeyByID(ctx, platformKeyID)
	if err != nil {
		return "", err
	}
	if keyPtr == nil {
		return "", fmt.Errorf("platform key not found")
	}
	key := *keyPtr
	// DepartmentID is pre-written by upsertPendingPlatformKeyMapping before TrySyncCreate is called.
	// At this point existing is guaranteed non-nil (pending or synced mapping exists).
	if existing == nil {
		return "", fmt.Errorf("platform key mapping missing for %s", platformKeyID)
	}
	departmentID := existing.DepartmentID
	if departmentID == uuid.Nil {
		return "", fmt.Errorf("department not resolved for key")
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

	deptAllowed := common.ResolveDeptAllowedModelIDs(departmentID, departments, rules, models)
	_, effectiveCallTypes := resolveModelLimits(d, models, key.ModelWhitelist, deptAllowed)

	group := d.ChannelPolicy.ResolveNewAPIGroup(ctx, departmentID)
	displayName := departmentID.String()
	if dept := pkgorg.FindDepartment(departments, departmentID); dept != nil {
		displayName = dept.Name
	}
	if err := d.Client.EnsureGroup(ctx, group, displayName); err != nil {
		return "", fmt.Errorf("ensure newapi group %s: %w", group, err)
	}

	walletCompanyID, ok := company.ResolveNewAPIWalletCompanyID(ctx, d.Store.Company())
	if !ok || walletCompanyID <= 0 {
		return "", fmt.Errorf("newapi wallet company id required for platform key %s", key.ID)
	}
	req := adminport.CreateTokenInput{
		UserID:             walletCompanyID,
		Name:               TokenName(key.ID),
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
		deleteRemoteTokenBestEffort(ctx, d, key.ID, token.ID)
		return "", err
	}
	now := time.Now()
	if err := d.Mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, now); err != nil {
		deleteRemoteTokenBestEffort(ctx, d, key.ID, token.ID)
		return "", err
	}
	return token.Key, nil
}

func deleteRemoteTokenBestEffort(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID, tokenID int64) {
	if tokenID <= 0 {
		return
	}
	if err := d.Client.DeleteToken(ctx, tokenID); err != nil {
		slog.Default().Warn("compensate delete newapi token failed",
			"platform_key_id", platformKeyID, "newapi_token_id", tokenID, "error", err)
	}
}

func RollbackFailedCreate(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) {
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
func SyncPlatformKeyCreate(ctx context.Context, d syncdeps.Deps, key types.PlatformKey, departmentID uuid.UUID) (string, error) {
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
