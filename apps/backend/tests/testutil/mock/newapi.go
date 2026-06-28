package mock

import (
	"context"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type StubAdminClient struct {
	Token newapi.Token

	CreateTokenFn      func(ctx context.Context, req newapi.CreateTokenRequest) (newapi.Token, error)
	UpdateTokenFn      func(ctx context.Context, req newapi.UpdateTokenRequest) (newapi.Token, error)
	GetTokenFn         func(ctx context.Context, tokenID int64) (newapi.Token, error)
	DeleteTokenFn      func(ctx context.Context, tokenID int64) error
	UpsertChannelFn    func(ctx context.Context, req newapi.UpsertChannelRequest) (newapi.Channel, error)
	RebuildAbilitiesFn func(ctx context.Context) error
	ListLogsFn         func(ctx context.Context, params newapi.ListLogsParams) ([]newapi.LogEntry, error)

	CreateTokenCalls      int
	UpdateTokenCalls      int
	GetTokenCalls         int
	DeleteTokenCalls      int
	UpsertChannelCalls    int
	RebuildAbilitiesCalls int
	ListLogsCalls         int
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

func (s *StubAdminClient) DeleteToken(ctx context.Context, tokenID int64) error {
	s.DeleteTokenCalls++
	if s.DeleteTokenFn != nil {
		return s.DeleteTokenFn(ctx, tokenID)
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

func (s *StubAdminClient) ListLogs(ctx context.Context, params newapi.ListLogsParams) ([]newapi.LogEntry, error) {
	s.ListLogsCalls++
	if s.ListLogsFn != nil {
		return s.ListLogsFn(ctx, params)
	}
	return nil, nil
}
