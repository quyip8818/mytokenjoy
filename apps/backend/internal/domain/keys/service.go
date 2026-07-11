package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	CreateProviderKeyForPlatform(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	ToggleProviderKey(ctx context.Context, id string, enabled bool) error
	RotateProviderKey(ctx context.Context, id string, newKey string) (types.ProviderKey, error)
	DeleteProviderKey(ctx context.Context, id string) error
	ListPlatformKeys(ctx context.Context, filter types.PlatformKeyListFilter) (types.PageResult[types.PlatformKey], error)
	BudgetSummary(ctx context.Context, memberID string) (types.MemberBudgetSummary, error)
	CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error)
	UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error)
	TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error)
	RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error)
	RevokePlatformKey(ctx context.Context, id string) error
	DeletePlatformKey(ctx context.Context, id string) error
	ListApprovals(ctx context.Context, tab, memberID string) ([]types.KeyApproval, error)
	CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error)
	ApprovalBudgetCheck(ctx context.Context, id string) (types.ApprovalBudgetCheck, error)
	ApproveApproval(ctx context.Context, id string, approverMemberID string) error
	RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error
}

type service struct {
	cfg        config.Config
	store      store.Store
	delayer    common.Delayer
	newAPISync newapisync.KeysNewAPISync
}

func NewService(cfg config.Config, st store.Store, newAPISync newapisync.KeysNewAPISync, delayer common.Delayer) Service {
	return &service{
		cfg:        cfg,
		store:      st,
		delayer:    delayer,
		newAPISync: newAPISync,
	}
}

func (s *service) ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	return s.store.Keys().ProviderKeys(ctx)
}

func (s *service) BudgetSummary(ctx context.Context, memberID string) (types.MemberBudgetSummary, error) {
	if memberID == "" {
		return types.MemberBudgetSummary{}, domain.BadRequest("memberId is required")
	}
	tree, err := budget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.Clock())
	if err != nil {
		return types.MemberBudgetSummary{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.MemberBudgetSummary{}, err
	}
	platformKeys, err := budget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.MemberBudgetSummary{}, err
	}
	reservedPool := budget.GetReservedPoolForMember(tree, members, memberID)
	return budget.BuildBudgetSummary(members, platformKeys, memberID, reservedPool), nil
}

func (s *service) ListApprovals(ctx context.Context, tab, memberID string) ([]types.KeyApproval, error) {
	items, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]types.KeyApproval, 0, len(items))
	for _, item := range items {
		switch tab {
		case "pending", "approved", "rejected":
			if item.Status != tab {
				continue
			}
		case "all", "":
			// no status filter
		default:
			if item.Status != tab {
				continue
			}
		}
		if memberID != "" && item.ApplicantID != memberID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (s *service) ApprovalBudgetCheck(ctx context.Context, id string) (types.ApprovalBudgetCheck, error) {
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return types.ApprovalBudgetCheck{}, err
	}
	var approval *types.KeyApproval
	for i := range approvals {
		if approvals[i].ID == id {
			approval = &approvals[i]
			break
		}
	}
	if approval == nil {
		return types.ApprovalBudgetCheck{}, domain.NotFound("Not found")
	}
	requested := approval.RequestedBudget
	tree, err := budget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.Clock())
	if err != nil {
		return types.ApprovalBudgetCheck{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.ApprovalBudgetCheck{}, err
	}
	reservedPool := budget.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	return types.ApprovalBudgetCheck{
		Sufficient: requested <= reservedPool, ReservedPool: reservedPool, Requested: requested,
	}, nil
}
