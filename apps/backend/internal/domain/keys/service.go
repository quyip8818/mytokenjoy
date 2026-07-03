package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	CreatePlatformProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	ToggleProviderKey(ctx context.Context, id string) error
	RotateProviderKey(ctx context.Context, id string) (types.ProviderKey, error)
	DeleteProviderKey(ctx context.Context, id string) error
	ListPlatformKeys(ctx context.Context, memberID, budgetGroupID string) (types.PageResult[types.PlatformKey], error)
	QuotaSummary(ctx context.Context, memberID string) (types.MemberQuotaSummary, error)
	CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error)
	UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error)
	TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error)
	RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error)
	RevokePlatformKey(ctx context.Context, id string) error
	DeletePlatformKey(ctx context.Context, id string) error
	ListApprovals(ctx context.Context, tab, memberID string) ([]types.KeyApproval, error)
	CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error)
	ApprovalQuotaCheck(ctx context.Context, id string) (types.ApprovalQuotaCheck, error)
	ApproveApproval(ctx context.Context, id string, approverMemberID string) error
	RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error
}

type service struct {
	cfg       config.Config
	store     store.Store
	delayer   common.Delayer
	relaySync relay.KeysRelaySync
}

func NewService(cfg config.Config, st store.Store, relaySync relay.KeysRelaySync, delayer common.Delayer) Service {
	return &service{
		cfg:       cfg,
		store:     st,
		delayer:   delayer,
		relaySync: relaySync,
	}
}

func (s *service) ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	return s.store.Keys().ProviderKeys(ctx)
}

func (s *service) ListPlatformKeys(ctx context.Context, memberID, budgetGroupID string) (types.PageResult[types.PlatformKey], error) {
	items, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PageResult[types.PlatformKey]{}, err
	}
	filtered := make([]types.PlatformKey, 0, len(items))
	for _, key := range items {
		if memberID != "" && (key.MemberID == nil || *key.MemberID != memberID) {
			continue
		}
		if budgetGroupID != "" && (key.BudgetGroupID == nil || *key.BudgetGroupID != budgetGroupID) {
			continue
		}
		filtered = append(filtered, key)
	}
	return types.PageResult[types.PlatformKey]{
		Items: filtered, Total: len(filtered), Page: 1, PageSize: 20,
	}, nil
}

func (s *service) QuotaSummary(ctx context.Context, memberID string) (types.MemberQuotaSummary, error) {
	if memberID == "" {
		return types.MemberQuotaSummary{}, domain.BadRequest("memberId is required")
	}
	tree, err := common.LoadBudgetTree(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.MemberQuotaSummary{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.MemberQuotaSummary{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.MemberQuotaSummary{}, err
	}
	reservedPool := budget.GetReservedPoolForMember(tree, members, memberID)
	return budget.BuildQuotaSummary(members, platformKeys, memberID, reservedPool), nil
}

func (s *service) ListApprovals(ctx context.Context, tab, memberID string) ([]types.KeyApproval, error) {
	items, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]types.KeyApproval, 0, len(items))
	for _, item := range items {
		if tab == "pending" && item.Status != "pending" {
			continue
		}
		if tab == "mine" && memberID != "" && item.ApplicantID != memberID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (s *service) ApprovalQuotaCheck(ctx context.Context, id string) (types.ApprovalQuotaCheck, error) {
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return types.ApprovalQuotaCheck{}, err
	}
	var approval *types.KeyApproval
	for i := range approvals {
		if approvals[i].ID == id {
			approval = &approvals[i]
			break
		}
	}
	requested := 0.0
	reservedPool := 0.0
	if approval != nil {
		requested = approval.RequestedQuota
		tree, err := common.LoadBudgetTree(ctx, s.store.Org().Nodes())
		if err != nil {
			return types.ApprovalQuotaCheck{}, err
		}
		members, err := s.store.Org().Members(ctx)
		if err != nil {
			return types.ApprovalQuotaCheck{}, err
		}
		reservedPool = budget.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	}
	return types.ApprovalQuotaCheck{
		Sufficient: requested <= reservedPool, ReservedPool: reservedPool, Requested: requested,
	}, nil
}
