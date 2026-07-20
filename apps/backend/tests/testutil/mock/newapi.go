package mock

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

// StubAdminClient implements adminport.Port for testing.
type StubAdminClient struct {
	Token newapi.Token
	User  newapi.User

	CreateTokenFn      func(ctx context.Context, req adminport.CreateTokenInput) (adminport.TokenResult, error)
	UpdateTokenFn      func(ctx context.Context, req adminport.UpdateTokenInput) (adminport.TokenResult, error)
	GetTokenFn         func(ctx context.Context, tokenID int64) (adminport.TokenResult, error)
	GetTokenKeyFn      func(ctx context.Context, tokenID int64) (string, error)
	DeleteTokenFn      func(ctx context.Context, tokenID int64) error
	RegenerateTokenFn  func(ctx context.Context, tokenID int64) (adminport.TokenResult, error)
	CreateUserFn       func(ctx context.Context, req adminport.CreateUserInput) (adminport.UserResult, error)
	GetUserQuotaFn     func(ctx context.Context, userID int64) (int64, error)
	TopUpFn            func(ctx context.Context, req adminport.TopUpInput) error
	UpsertChannelFn    func(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error)
	RebuildAbilitiesFn func(ctx context.Context) error
	EnsureGroupFn      func(ctx context.Context, group, displayName string) error
	ListModelPricingFn func(ctx context.Context) ([]adminport.ModelPricing, error)

	CreateTokenCalls      int
	UpdateTokenCalls      int
	GetTokenCalls         int
	GetTokenKeyCalls      int
	DeleteTokenCalls      int
	RegenerateTokenCalls  int
	CreateUserCalls       int
	GetUserQuotaCalls     int
	TopUpCalls            int
	UpsertChannelCalls    int
	RebuildAbilitiesCalls int
	EnsureGroupCalls      int
	ListModelPricingCalls int
}

func (s *StubAdminClient) defaultTokenResult() adminport.TokenResult {
	return adminport.TokenResult{
		ID:          s.Token.ID,
		UserID:      s.Token.UserID,
		Key:         s.Token.Key,
		RemainQuota: s.Token.RemainQuota,
		Group:       s.Token.Group,
	}
}

func (s *StubAdminClient) CreateToken(ctx context.Context, req adminport.CreateTokenInput) (adminport.TokenResult, error) {
	s.CreateTokenCalls++
	if s.CreateTokenFn != nil {
		return s.CreateTokenFn(ctx, req)
	}
	return s.defaultTokenResult(), nil
}

func (s *StubAdminClient) UpdateToken(ctx context.Context, req adminport.UpdateTokenInput) (adminport.TokenResult, error) {
	s.UpdateTokenCalls++
	if s.UpdateTokenFn != nil {
		return s.UpdateTokenFn(ctx, req)
	}
	return s.defaultTokenResult(), nil
}

func (s *StubAdminClient) GetToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	s.GetTokenCalls++
	if s.GetTokenFn != nil {
		return s.GetTokenFn(ctx, tokenID)
	}
	return s.defaultTokenResult(), nil
}

func (s *StubAdminClient) GetTokenKey(ctx context.Context, tokenID int64) (string, error) {
	s.GetTokenKeyCalls++
	if s.GetTokenKeyFn != nil {
		return s.GetTokenKeyFn(ctx, tokenID)
	}
	if s.Token.Key != "" {
		return s.Token.Key, nil
	}
	return "", nil
}

func (s *StubAdminClient) DeleteToken(ctx context.Context, tokenID int64) error {
	s.DeleteTokenCalls++
	if s.DeleteTokenFn != nil {
		return s.DeleteTokenFn(ctx, tokenID)
	}
	return nil
}

func (s *StubAdminClient) RegenerateToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	s.RegenerateTokenCalls++
	if s.RegenerateTokenFn != nil {
		return s.RegenerateTokenFn(ctx, tokenID)
	}
	return s.defaultTokenResult(), nil
}

func (s *StubAdminClient) CreateUser(ctx context.Context, req adminport.CreateUserInput) (adminport.UserResult, error) {
	s.CreateUserCalls++
	if s.CreateUserFn != nil {
		return s.CreateUserFn(ctx, req)
	}
	return adminport.UserResult{ID: s.User.ID}, nil
}

func (s *StubAdminClient) GetUserQuota(ctx context.Context, userID int64) (int64, error) {
	s.GetUserQuotaCalls++
	if s.GetUserQuotaFn != nil {
		return s.GetUserQuotaFn(ctx, userID)
	}
	return s.User.Quota, nil
}

func (s *StubAdminClient) TopUp(ctx context.Context, req adminport.TopUpInput) error {
	s.TopUpCalls++
	if s.TopUpFn != nil {
		return s.TopUpFn(ctx, req)
	}
	return nil
}

func (s *StubAdminClient) UpsertChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	s.UpsertChannelCalls++
	if s.UpsertChannelFn != nil {
		return s.UpsertChannelFn(ctx, req)
	}
	return adminport.ChannelResult{}, nil
}

func (s *StubAdminClient) RebuildAbilities(ctx context.Context) error {
	s.RebuildAbilitiesCalls++
	if s.RebuildAbilitiesFn != nil {
		return s.RebuildAbilitiesFn(ctx)
	}
	return nil
}

func (s *StubAdminClient) EnsureGroup(ctx context.Context, group, displayName string) error {
	s.EnsureGroupCalls++
	if s.EnsureGroupFn != nil {
		return s.EnsureGroupFn(ctx, group, displayName)
	}
	return nil
}

func (s *StubAdminClient) ListModelPricing(ctx context.Context) ([]adminport.ModelPricing, error) {
	s.ListModelPricingCalls++
	if s.ListModelPricingFn != nil {
		return s.ListModelPricingFn(ctx)
	}
	return nil, nil
}

var _ adminport.Port = (*StubAdminClient)(nil)
