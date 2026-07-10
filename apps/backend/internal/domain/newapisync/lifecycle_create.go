package newapisync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID string) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("newapi not enabled")
	}
	mapping := store.PlatformKeyMapping{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
		MemberID:      key.MemberID,
		DepartmentID:  departmentID,
		BudgetGroupID: key.BudgetGroupID,
		NewAPIGroup:   l.channelPolicy.ResolveNewAPIGroup(ctx, departmentID),
		SyncStatus:    store.MappingSyncStatusPending,
	}
	if err := l.mappings.UpsertMapping(ctx, mapping); err != nil {
		return err
	}
	payload, _ := json.Marshal(CreateKeyOutboxPayload{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
	})
	return l.outbox.EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID:      fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:    store.OutboxKindCreateKey,
		Payload: payload,
		Status:  store.JobStatusPending,
	})
}

func (l *NewAPISync) TrySyncCreate(ctx context.Context, platformKeyID string) (string, error) {
	if !l.Enabled() {
		return "", domain.ServiceUnavailable("newapi not enabled")
	}
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.store.Keys(), l.cfg.Clock())
	if err != nil {
		return "", err
	}
	key, ok := findPlatformKey(platformKeys, platformKeyID)
	if !ok {
		return "", fmt.Errorf("platform key not found")
	}
	members, err := l.store.Org().Members(ctx)
	if err != nil {
		return "", err
	}
	departments, err := common.LoadDepartments(ctx, l.store.Org().Nodes())
	if err != nil {
		return "", err
	}
	rules, err := common.LoadRoutingRules(ctx, l.store.Org().Nodes(), l.store.Models().Allowlist())
	if err != nil {
		return "", err
	}
	models, err := l.store.Models().Models(ctx)
	if err != nil {
		return "", err
	}
	groups, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.cfg.Clock())
	if err != nil {
		return "", err
	}
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org().Nodes(), l.cfg.Clock())
	if err != nil {
		return "", err
	}

	departmentID := ""
	if key.MemberID != nil {
		if member, found := org.FindMemberByID(members, *key.MemberID); found {
			departmentID = member.DepartmentID
		}
	}
	if departmentID == "" && key.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *key.BudgetGroupID && len(group.DepartmentIDs) > 0 {
				departmentID = group.DepartmentIDs[0]
				break
			}
		}
	}
	if departmentID == "" {
		return "", fmt.Errorf("department not resolved for key")
	}

	deptAllowed := common.ResolveDeptAllowedModelIDs(departmentID, departments, rules, models)
	effectiveIDs := newapi.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	effectiveCallTypes := newapi.EffectiveCallTypes(models, effectiveIDs)
	remainCNY := ComputeRemainQuota(key, tree, members, platformKeys, groups, departmentID)
	remainUnits := l.capRemainUnits(ctx, remainCNY, models, effectiveIDs)

	walletUserID := l.newAPIWalletUserID(ctx)
	req := newapi.CreateTokenRequest{
		UserID:             walletUserID,
		Name:               key.Name,
		RemainQuota:        remainUnits,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: len(effectiveCallTypes) > 0,
		ModelLimits:        newapi.FormatModelLimits(effectiveCallTypes),
		Group:              l.channelPolicy.ResolveNewAPIGroup(ctx, departmentID),
		ExpiredTime:        -1,
	}
	token, err := l.client.CreateToken(ctx, req)
	if err != nil {
		return "", err
	}
	now := time.Now()
	remain := token.RemainQuota
	if err := l.mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.MappingSyncStatusSynced, &remain, now); err != nil {
		return "", err
	}
	keys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return "", err
	}
	for i := range keys {
		if keys[i].ID == key.ID {
			keys[i].FullKey = &token.Key
			keys[i].KeyPrefix = newAPIPlatformKeyPrefix(token.Key)
			if err := l.store.Keys().SetPlatformKeys(ctx, keys); err != nil {
				return "", err
			}
			break
		}
	}
	return token.Key, nil
}

func (l *NewAPISync) RollbackFailedCreate(ctx context.Context, platformKeyID string) {
	keys, err := l.store.Keys().PlatformKeys(ctx)
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
	if err := l.store.Keys().SetPlatformKeys(ctx, filtered); err != nil {
		slog.Default().Warn("rollback failed create persist failed", "platform_key_id", platformKeyID, "error", err)
	}
}
