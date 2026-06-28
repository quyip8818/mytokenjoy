package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetlookup"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListProviderKeys() []types.ProviderKey
	CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	ToggleProviderKey(ctx context.Context, id string) error
	RotateProviderKey(ctx context.Context, id string) (types.ProviderKey, error)
	DeleteProviderKey(id string) error
	ListPlatformKeys(memberID, budgetGroupID string) types.PageResult[types.PlatformKey]
	QuotaSummary(memberID string) types.MemberQuotaSummary
	CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error)
	UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error)
	TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error)
	RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error)
	RevokePlatformKey(ctx context.Context, id string) error
	DeletePlatformKey(id string) error
	ListApprovals(tab, memberID string) []types.KeyApproval
	CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error)
	ApprovalQuotaCheck(id string) types.ApprovalQuotaCheck
	ApproveApproval(ctx context.Context, id string, approverMemberID string) error
	RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error
}

type service struct {
	cfg       config.Config
	store     store.Store
	delayer   simulate.Delayer
	lifecycle relay.Lifecycle
}

func NewService(cfg config.Config, st store.Store, lifecycle relay.Lifecycle) Service {
	return &service{
		cfg:       cfg,
		store:     st,
		delayer:   simulate.NewDelayer(cfg.SimulateDelay),
		lifecycle: lifecycle,
	}
}

func (s *service) ListProviderKeys() []types.ProviderKey {
	return s.store.Keys().ProviderKeys()
}

func (s *service) ListPlatformKeys(memberID, budgetGroupID string) types.PageResult[types.PlatformKey] {
	items := s.store.Keys().PlatformKeys()
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
	}
}

func (s *service) QuotaSummary(memberID string) types.MemberQuotaSummary {
	if memberID == "" {
		memberID = "m-1"
	}
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	reservedPool := budgetlookup.GetReservedPoolForMember(tree, members, memberID)
	return memberquota.BuildQuotaSummary(pools, platformKeys, memberID, reservedPool)
}

func (s *service) ListApprovals(tab, memberID string) []types.KeyApproval {
	items := s.store.Keys().Approvals()
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
	return filtered
}

func (s *service) ApprovalQuotaCheck(id string) types.ApprovalQuotaCheck {
	approvals := s.store.Keys().Approvals()
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
		tree := s.store.Budget().Tree()
		members := s.store.Org().Members()
		reservedPool = budgetlookup.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	}
	return types.ApprovalQuotaCheck{
		Sufficient: requested <= reservedPool, ReservedPool: reservedPool, Requested: requested,
	}
}
