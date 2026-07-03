package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) EnqueueUpdatePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return nil
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

func (l *TokenLifecycle) SyncUpdatePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return fmt.Errorf("relay mapping missing for %s", platformKeyID)
	}
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
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
	groups, err := l.store.Budget().Groups(ctx)
	if err != nil {
		return err
	}
	tree, err := common.LoadBudgetTree(ctx, l.store.Org().Nodes())
	if err != nil {
		return err
	}

	deptAllowed := common.ResolveDeptAllowedModels(mapping.DepartmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	remainCNY := ComputeRemainQuotaCNY(key, tree, members, platformKeys, groups, mapping.DepartmentID)
	remainUnits := l.capRemainUnits(ctx, remainCNY, models, effective)
	status := newapi.TokenStatusEnabled
	if key.Status != "active" {
		status = newapi.TokenStatusDisabled
	}
	remain := remainUnits
	enabled := len(effective) > 0
	req := newapi.UpdateTokenRequest{
		ID:                 *mapping.NewAPITokenID,
		Name:               key.Name,
		Status:             &status,
		RemainQuota:        &remain,
		ModelLimitsEnabled: &enabled,
		ModelLimits:        newapi.FormatModelLimits(effective),
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

func (l *TokenLifecycle) SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return nil
	}
	if err := l.client.DeleteToken(ctx, *mapping.NewAPITokenID); err != nil {
		return err
	}
	mapping.SyncStatus = store.RelaySyncStatusSynced
	return l.mappings.UpsertMapping(ctx, *mapping)
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
