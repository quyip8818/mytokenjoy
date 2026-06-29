package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/store"
)

type TokenLifecycle struct {
	cfg    config.Config
	store  store.Store
	client newapi.AdminClient
}

func NewTokenLifecycle(cfg config.Config, st store.Store, client newapi.AdminClient) *TokenLifecycle {
	return &TokenLifecycle{cfg: cfg, store: st, client: client}
}

var _ Lifecycle = (*TokenLifecycle)(nil)

func (l *TokenLifecycle) Enabled() bool {
	return l.cfg.NewAPIEnabled && l.client != nil
}

func (l *TokenLifecycle) SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	mapping := store.RelayMapping{
		PlatformKeyID: key.ID,
		MemberID:      key.MemberID,
		DepartmentID:  departmentID,
		BudgetGroupID: key.BudgetGroupID,
		RelayGroup:    newapi.RelayGroupForDepartment(departmentID),
		SyncStatus:    store.RelaySyncStatusPending,
	}
	if err := l.store.Relay().UpsertMapping(mapping); err != nil {
		return err
	}
	payload, _ := json.Marshal(CreateTokenOutboxPayload{PlatformKeyID: key.ID})
	return l.store.Relay().EnqueueRelayOutbox(store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindCreateToken,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) TrySyncCreate(ctx context.Context, platformKeyID string) (string, error) {
	if !l.Enabled() {
		return "", nil
	}
	key, ok := findPlatformKey(l.store.Keys().PlatformKeys(), platformKeyID)
	if !ok {
		return "", fmt.Errorf("platform key not found")
	}
	members := l.store.Org().Members()
	departments := l.store.Org().Departments()
	rules := l.store.Models().RoutingRules()
	models := l.store.Models().Models()
	pools := l.store.Budget().MemberQuotaPools()
	platformKeys := l.store.Keys().PlatformKeys()
	groups := l.store.Budget().Groups()
	tree := l.store.Budget().Tree()

	departmentID := ""
	if key.MemberID != nil {
		if member, found := queryutil.FindMemberByID(members, *key.MemberID); found {
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

	deptAllowed := routingutil.ResolveDeptAllowedModels(departmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	remainCNY := ComputeRemainQuotaCNY(key, tree, pools, platformKeys, groups, departmentID)
	remainUnits := newapi.ToNewAPIUnits(remainCNY, models, effective)

	req := newapi.CreateTokenRequest{
		Name:               key.Name,
		RemainQuota:        remainUnits,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: len(effective) > 0,
		ModelLimits:        newapi.FormatModelLimits(effective),
		Group:              newapi.RelayGroupForDepartment(departmentID),
		ExpiredTime:        -1,
	}
	token, err := l.client.CreateToken(ctx, req)
	if err != nil {
		return "", err
	}
	now := time.Now()
	remain := token.RemainQuota
	if err := l.store.Relay().UpdateMappingSync(key.ID, token.ID, store.RelaySyncStatusSynced, &remain, now); err != nil {
		return "", err
	}
	keys := l.store.Keys().PlatformKeys()
	for i := range keys {
		if keys[i].ID == key.ID {
			keys[i].FullKey = &token.Key
			prefix := token.Key
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			keys[i].KeyPrefix = prefix
			if err := l.store.Keys().SetPlatformKeys(keys); err != nil {
				return "", err
			}
			break
		}
	}
	return token.Key, nil
}

func (l *TokenLifecycle) EnqueueUpdatePlatformKey(platformKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpdateTokenOutboxPayload{PlatformKeyID: platformKeyID})
	return l.store.Relay().EnqueueRelayOutbox(store.RelayOutboxEntry{
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
	mapping, err := l.store.Relay().GetMappingByPlatformKeyID(platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return fmt.Errorf("relay mapping missing for %s", platformKeyID)
	}
	key, ok := findPlatformKey(l.store.Keys().PlatformKeys(), platformKeyID)
	if !ok {
		return fmt.Errorf("platform key not found")
	}
	departments := l.store.Org().Departments()
	rules := l.store.Models().RoutingRules()
	models := l.store.Models().Models()
	pools := l.store.Budget().MemberQuotaPools()
	platformKeys := l.store.Keys().PlatformKeys()
	groups := l.store.Budget().Groups()
	tree := l.store.Budget().Tree()

	deptAllowed := routingutil.ResolveDeptAllowedModels(mapping.DepartmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	remainCNY := ComputeRemainQuotaCNY(key, tree, pools, platformKeys, groups, mapping.DepartmentID)
	remainUnits := newapi.ToNewAPIUnits(remainCNY, models, effective)
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
	return l.store.Relay().UpdateMappingSync(key.ID, token.ID, store.RelaySyncStatusSynced, &relayRemain, now)
}

func (l *TokenLifecycle) SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	mapping, err := l.store.Relay().GetMappingByPlatformKeyID(platformKeyID)
	if err != nil || mapping == nil || mapping.NewAPITokenID == nil {
		return nil
	}
	if err := l.client.DeleteToken(ctx, *mapping.NewAPITokenID); err != nil {
		return err
	}
	mapping.SyncStatus = store.RelaySyncStatusSynced
	return l.store.Relay().UpsertMapping(*mapping)
}

func (l *TokenLifecycle) DisablePlatformKey(ctx context.Context, platformKeyID string) error {
	keys := l.store.Keys().PlatformKeys()
	for i := range keys {
		if keys[i].ID == platformKeyID {
			keys[i].Status = "disabled"
			if err := l.store.Keys().SetPlatformKeys(keys); err != nil {
				return err
			}
			break
		}
	}
	if !l.Enabled() {
		return nil
	}
	mapping, err := l.store.Relay().GetMappingByPlatformKeyID(platformKeyID)
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

func (l *TokenLifecycle) EnqueueUpsertProviderKey(providerKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpsertChannelOutboxPayload{ProviderKeyID: providerKeyID})
	return l.store.Relay().EnqueueRelayOutbox(store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindUpsertChannel,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) SyncUpsertProviderKey(ctx context.Context, providerKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	keys := l.store.Keys().ProviderKeys()
	idx := -1
	for i := range keys {
		if keys[i].ID == providerKeyID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("provider key not found: %s", providerKeyID)
	}
	pk := keys[idx]
	if pk.SecretKey == "" {
		return fmt.Errorf("provider key secret missing: %s", providerKeyID)
	}
	status := newapi.ChannelStatusEnabled
	if pk.Status != "active" {
		status = newapi.ChannelStatusDisabled
	}
	req := newapi.UpsertChannelRequest{
		Type:   newapi.ProviderChannelType(pk.Provider),
		Name:   pk.Name,
		Key:    pk.SecretKey,
		Status: status,
	}
	if pk.RelayChannelID > 0 {
		req.ID = pk.RelayChannelID
	}
	channel, err := l.client.UpsertChannel(ctx, req)
	if err != nil {
		return err
	}
	keys[idx].RelayChannelID = channel.ID
	return l.store.Keys().SetProviderKeys(keys)
}

func (l *TokenLifecycle) EnqueueModelLimitsForDepartment(departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpdateModelLimitsOutboxPayload{DepartmentID: departmentID})
	return l.store.Relay().EnqueueRelayOutbox(store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindUpdateModelLimits,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) EnqueueModelLimitsForDepartments(departmentIDs []string) error {
	seen := make(map[string]struct{}, len(departmentIDs))
	for _, deptID := range departmentIDs {
		if deptID == "" {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		seen[deptID] = struct{}{}
		if err := l.EnqueueModelLimitsForDepartment(deptID); err != nil {
			return err
		}
	}
	return nil
}

func (l *TokenLifecycle) SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	mappings, err := l.store.Relay().ListMappingsByDepartmentID(departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.SyncStatus != store.RelaySyncStatusSynced || mapping.NewAPITokenID == nil {
			continue
		}
		if err := l.SyncUpdatePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

func (l *TokenLifecycle) EnqueueRebalanceAxis(axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(RebalanceAxisOutboxPayload{AxisKind: axisKind, AxisID: axisID})
	return l.store.Relay().EnqueueRelayOutbox(store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindRebalanceToken,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) RollbackFailedCreate(platformKeyID string) {
	keys := l.store.Keys().PlatformKeys()
	filtered := make([]types.PlatformKey, 0, len(keys))
	for _, key := range keys {
		if key.ID != platformKeyID {
			filtered = append(filtered, key)
		}
	}
	if err := l.store.Keys().SetPlatformKeys(filtered); err != nil {
		slog.Default().Warn("rollback failed create persist failed", "platform_key_id", platformKeyID, "error", err)
	}
}

func findPlatformKey(keys []types.PlatformKey, id string) (types.PlatformKey, bool) {
	for _, key := range keys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}
