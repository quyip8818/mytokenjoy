package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) EnqueueUpdatePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("relay not enabled")
	}
	payload, _ := json.Marshal(UpdateTokenOutboxPayload{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: platformKeyID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindUpdateToken,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) SyncUpdatePlatformKey(ctx context.Context, platformKeyID string, targetActive *bool) error {
	if !l.Enabled() {
		return domain.ServiceUnavailable("relay not enabled")
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return fmt.Errorf("relay mapping missing for %s", platformKeyID)
	}
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.store.Keys(), l.cfg.NowUTC())
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
	groups, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org(), l.store.Budget(), l.cfg.NowUTC())
	if err != nil {
		return err
	}
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, l.store.BudgetSnapshots(), l.store.Org().Nodes(), l.cfg.NowUTC())
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
		ID:                 *mapping.NewAPITokenID,
		Name:               key.Name,
		Status:             &status,
		RemainQuota:        &remain,
		ModelLimitsEnabled: &enabled,
		ModelLimits:        newapi.FormatModelLimits(effectiveCallTypes),
		Group:              mapping.RelayGroup,
	}
	token, err := l.client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	now := time.Now()
	relayRemain := token.RemainQuota
	return l.mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.RelaySyncStatusSynced, &relayRemain, now)
}

func (l *TokenLifecycle) DisablePlatformKey(ctx context.Context, platformKeyID string) error {
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
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return nil
	}
	status := newapi.TokenStatusDisabled
	zero := int64(0)
	req := newapi.UpdateTokenRequest{
		ID:          *mapping.NewAPITokenID,
		Status:      &status,
		RemainQuota: &zero,
	}
	_, err = l.client.UpdateToken(ctx, req)
	return err
}
