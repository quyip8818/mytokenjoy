package budget

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/store"
)

type Rebalancer interface {
	ProcessAxis(ctx context.Context, axisKind, axisID string) error
}

type RebalanceService struct {
	cfg       config.Config
	store     store.Store
	client    newapi.AdminClient
	lifecycle relay.Lifecycle
}

func NewRebalanceService(cfg config.Config, st store.Store, client newapi.AdminClient, lifecycle relay.Lifecycle) *RebalanceService {
	return &RebalanceService{cfg: cfg, store: st, client: client, lifecycle: lifecycle}
}

func (s *RebalanceService) ProcessAxis(ctx context.Context, axisKind, axisID string) error {
	if s.client == nil || !s.cfg.NewAPIEnabled {
		return nil
	}
	var mappings []store.RelayMapping
	var err error
	switch axisKind {
	case store.RebalanceAxisMember:
		mappings, err = s.store.Relay().ListMappingsByMemberID(axisID)
	case store.RebalanceAxisDepartment:
		mappings, err = s.store.Relay().ListMappingsByDepartmentID(axisID)
	case store.RebalanceAxisBudgetGroup:
		mappings, err = s.store.Relay().ListMappingsByBudgetGroupID(axisID)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.NewAPITokenID == nil || mapping.SyncStatus != store.RelaySyncStatusSynced {
			continue
		}
		if err := s.rebalanceKey(ctx, mapping); err != nil {
			return err
		}
	}
	return nil
}

func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.RelayMapping) error {
	platformKeys := s.store.Keys().PlatformKeys()
	key, ok := findPlatformKeyByID(platformKeys, mapping.PlatformKeyID)
	if !ok || key.Status != "active" {
		return nil
	}
	token, err := s.client.GetToken(ctx, *mapping.NewAPITokenID)
	if err != nil {
		return err
	}

	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	pools := s.store.Budget().MemberQuotaPools()
	groups := s.store.Budget().Groups()
	tree := s.store.Budget().Tree()

	deptAllowed := routingutil.ResolveDeptAllowedModels(mapping.DepartmentID, departments, rules, models)
	effective := newapi.EffectiveWhitelist(key.ModelWhitelist, deptAllowed)
	newRemain := newapi.ToNewAPIUnits(
		relay.ComputeRemainQuotaCNY(key, tree, pools, platformKeys, groups, mapping.DepartmentID),
		models,
		effective,
	)
	if newRemain >= token.RemainQuota {
		return nil
	}
	remain := newRemain
	req := newapi.UpdateTokenRequest{
		ID:          token.ID,
		RemainQuota: &remain,
	}
	updated, err := s.client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	return s.store.Relay().UpdateMappingRemainQuota(mapping.PlatformKeyID, updated.RemainQuota)
}

func findPlatformKeyByID(platformKeys []types.PlatformKey, id string) (types.PlatformKey, bool) {
	for _, key := range platformKeys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}

var _ Rebalancer = (*RebalanceService)(nil)
