package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type TokenLifecycle struct {
	cfg           config.Config
	store         store.Store
	client        newapi.AdminClient
	mappings      store.RelayMappingRepository
	relayOutbox   store.RelayOutboxRepository
	wallet        company.WalletService
	channelPolicy ChannelPolicy
}

func NewTokenLifecycle(cfg config.Config, st store.Store, client newapi.AdminClient, wallet company.WalletService, channelPolicy ChannelPolicy) *TokenLifecycle {
	relayRepo := st.Relay()
	return &TokenLifecycle{
		cfg:           cfg,
		store:         st,
		client:        client,
		mappings:      relayRepo,
		relayOutbox:   relayRepo,
		wallet:        wallet,
		channelPolicy: channelPolicy,
	}
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
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
		MemberID:      key.MemberID,
		DepartmentID:  departmentID,
		BudgetGroupID: key.BudgetGroupID,
		RelayGroup:    l.channelPolicy.ResolveRelayGroup(ctx, departmentID),
		SyncStatus:    store.RelaySyncStatusPending,
	}
	if err := l.mappings.UpsertMapping(ctx, mapping); err != nil {
		return err
	}
	payload, _ := json.Marshal(CreateTokenOutboxPayload{
		CompanyID:     company.CompanyID(ctx),
		PlatformKeyID: key.ID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
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
	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
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
	departments, err := l.store.Org().Departments(ctx)
	if err != nil {
		return "", err
	}
	rules, err := l.store.Models().RoutingRules(ctx)
	if err != nil {
		return "", err
	}
	models, err := l.store.Models().Models(ctx)
	if err != nil {
		return "", err
	}
	pools, err := l.store.Budget().MemberQuotaPools(ctx)
	if err != nil {
		return "", err
	}
	groups, err := l.store.Budget().Groups(ctx)
	if err != nil {
		return "", err
	}
	tree, err := l.store.Budget().Tree(ctx)
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

	deptAllowed := common.ResolveDeptAllowedModels(departmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	remainCNY := ComputeRemainQuotaCNY(key, tree, pools, platformKeys, groups, departmentID)
	remainUnits := l.capRemainUnits(ctx, remainCNY, models, effective)

	walletUserID := l.walletUserID(ctx)
	req := newapi.CreateTokenRequest{
		UserID:             walletUserID,
		Name:               key.Name,
		RemainQuota:        remainUnits,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: len(effective) > 0,
		ModelLimits:        newapi.FormatModelLimits(effective),
		Group:              l.channelPolicy.ResolveRelayGroup(ctx, departmentID),
		ExpiredTime:        -1,
	}
	token, err := l.client.CreateToken(ctx, req)
	if err != nil {
		return "", err
	}
	now := time.Now()
	remain := token.RemainQuota
	if err := l.mappings.UpdateMappingSync(ctx, key.ID, token.ID, store.RelaySyncStatusSynced, &remain, now); err != nil {
		return "", err
	}
	keys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return "", err
	}
	for i := range keys {
		if keys[i].ID == key.ID {
			keys[i].FullKey = &token.Key
			prefix := token.Key
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			keys[i].KeyPrefix = prefix
			if err := l.store.Keys().SetPlatformKeys(ctx, keys); err != nil {
				return "", err
			}
			break
		}
	}
	return token.Key, nil
}

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
	departments, err := l.store.Org().Departments(ctx)
	if err != nil {
		return err
	}
	rules, err := l.store.Models().RoutingRules(ctx)
	if err != nil {
		return err
	}
	models, err := l.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	pools, err := l.store.Budget().MemberQuotaPools(ctx)
	if err != nil {
		return err
	}
	groups, err := l.store.Budget().Groups(ctx)
	if err != nil {
		return err
	}
	tree, err := l.store.Budget().Tree(ctx)
	if err != nil {
		return err
	}

	deptAllowed := common.ResolveDeptAllowedModels(mapping.DepartmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	remainCNY := ComputeRemainQuotaCNY(key, tree, pools, platformKeys, groups, mapping.DepartmentID)
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

func (l *TokenLifecycle) EnqueueUpsertProviderKey(ctx context.Context, providerKeyID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpsertChannelOutboxPayload{
		CompanyID:     company.CompanyID(ctx),
		ProviderKeyID: providerKeyID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
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
	keys, err := l.store.Keys().ProviderKeys(ctx)
	if err != nil {
		return err
	}
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
	return l.store.Keys().SetProviderKeys(ctx, keys)
}

func (l *TokenLifecycle) EnqueueModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpdateModelLimitsOutboxPayload{
		CompanyID:    company.CompanyID(ctx),
		DepartmentID: departmentID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindUpdateModelLimits,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error {
	seen := make(map[string]struct{}, len(departmentIDs))
	for _, deptID := range departmentIDs {
		if deptID == "" {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		seen[deptID] = struct{}{}
		if err := l.EnqueueModelLimitsForDepartment(ctx, deptID); err != nil {
			return err
		}
	}
	return nil
}

func (l *TokenLifecycle) SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	mappings, err := l.mappings.ListMappingsByDepartmentID(ctx, departmentID)
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

func (l *TokenLifecycle) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(RebalanceAxisOutboxPayload{
		CompanyID: company.CompanyID(ctx),
		AxisKind:  axisKind,
		AxisID:    axisID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindRebalanceToken,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) RollbackFailedCreate(ctx context.Context, platformKeyID string) {
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

func (l *TokenLifecycle) walletUserID(ctx context.Context) int64 {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletAccountID > 0 {
		return companyCtx.NewAPIWalletAccountID
	}
	companyID := company.CompanyID(ctx)
	company, err := l.store.Company().GetByID(ctx, companyID)
	if err != nil || company == nil || company.NewAPIWalletAccountID == nil {
		return 0
	}
	return *company.NewAPIWalletAccountID
}

func (l *TokenLifecycle) capRemainUnits(ctx context.Context, remainCNY float64, models []types.ModelInfo, effective []string) int64 {
	allocated := newapi.ToNewAPIUnits(remainCNY, models, effective)
	if l.wallet == nil {
		return allocated
	}
	walletID := l.walletUserID(ctx)
	if walletID <= 0 {
		return allocated
	}
	walletUnits, err := l.wallet.AvailableQuota(ctx, walletID)
	if err != nil {
		return allocated
	}
	if allocated < walletUnits {
		return allocated
	}
	return walletUnits
}

func findPlatformKey(keys []types.PlatformKey, id string) (types.PlatformKey, bool) {
	for _, key := range keys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}
