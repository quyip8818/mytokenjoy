package mock

import (
	"context"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type StubAdminClient struct {
	Token newapi.Token
	User  newapi.User

	CreateTokenFn      func(ctx context.Context, req newapi.CreateTokenRequest) (newapi.Token, error)
	UpdateTokenFn      func(ctx context.Context, req newapi.UpdateTokenRequest) (newapi.Token, error)
	GetTokenFn         func(ctx context.Context, tokenID int64) (newapi.Token, error)
	GetTokenKeyFn      func(ctx context.Context, tokenID int64) (string, error)
	DeleteTokenFn      func(ctx context.Context, tokenID int64) error
	RegenerateTokenFn  func(ctx context.Context, tokenID int64) (newapi.Token, error)
	CreateUserFn       func(ctx context.Context, req newapi.CreateUserRequest) (newapi.User, error)
	GetUserQuotaFn     func(ctx context.Context, userID int64) (int64, error)
	TopUpFn            func(ctx context.Context, req newapi.TopUpRequest) error
	UpsertChannelFn    func(ctx context.Context, req newapi.UpsertChannelRequest) (newapi.Channel, error)
	RebuildAbilitiesFn func(ctx context.Context) error

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
}

func (s *StubAdminClient) CreateToken(ctx context.Context, req newapi.CreateTokenRequest) (newapi.Token, error) {
	s.CreateTokenCalls++
	if s.CreateTokenFn != nil {
		return s.CreateTokenFn(ctx, req)
	}
	return s.Token, nil
}

func (s *StubAdminClient) UpdateToken(ctx context.Context, req newapi.UpdateTokenRequest) (newapi.Token, error) {
	s.UpdateTokenCalls++
	if s.UpdateTokenFn != nil {
		return s.UpdateTokenFn(ctx, req)
	}
	return s.Token, nil
}

func (s *StubAdminClient) GetToken(ctx context.Context, tokenID int64) (newapi.Token, error) {
	s.GetTokenCalls++
	if s.GetTokenFn != nil {
		return s.GetTokenFn(ctx, tokenID)
	}
	return s.Token, nil
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

func (s *StubAdminClient) RegenerateToken(ctx context.Context, tokenID int64) (newapi.Token, error) {
	s.RegenerateTokenCalls++
	if s.RegenerateTokenFn != nil {
		return s.RegenerateTokenFn(ctx, tokenID)
	}
	return s.Token, nil
}

func (s *StubAdminClient) CreateUser(ctx context.Context, req newapi.CreateUserRequest) (newapi.User, error) {
	s.CreateUserCalls++
	if s.CreateUserFn != nil {
		return s.CreateUserFn(ctx, req)
	}
	return s.User, nil
}

func (s *StubAdminClient) GetUserQuota(ctx context.Context, userID int64) (int64, error) {
	s.GetUserQuotaCalls++
	if s.GetUserQuotaFn != nil {
		return s.GetUserQuotaFn(ctx, userID)
	}
	return s.User.Quota, nil
}

func (s *StubAdminClient) TopUp(ctx context.Context, req newapi.TopUpRequest) error {
	s.TopUpCalls++
	if s.TopUpFn != nil {
		return s.TopUpFn(ctx, req)
	}
	return nil
}

func (s *StubAdminClient) UpsertChannel(ctx context.Context, req newapi.UpsertChannelRequest) (newapi.Channel, error) {
	s.UpsertChannelCalls++
	if s.UpsertChannelFn != nil {
		return s.UpsertChannelFn(ctx, req)
	}
	return newapi.Channel{}, nil
}

func (s *StubAdminClient) RebuildAbilities(ctx context.Context) error {
	s.RebuildAbilitiesCalls++
	if s.RebuildAbilitiesFn != nil {
		return s.RebuildAbilitiesFn(ctx)
	}
	return nil
}
