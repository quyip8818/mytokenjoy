package adminport

import "context"

type Port interface {
	CreateToken(ctx context.Context, req CreateTokenInput) (TokenResult, error)
	UpdateToken(ctx context.Context, req UpdateTokenInput) (TokenResult, error)
	GetToken(ctx context.Context, tokenID int64) (TokenResult, error)
	GetTokenKey(ctx context.Context, tokenID int64) (string, error)
	RegenerateToken(ctx context.Context, tokenID int64) (TokenResult, error)
	DeleteToken(ctx context.Context, tokenID int64) error
	UpsertChannel(ctx context.Context, req UpsertChannelInput) (ChannelResult, error)
	CreateUser(ctx context.Context, req CreateUserInput) (UserResult, error)
	TopUp(ctx context.Context, req TopUpInput) error
	RebuildAbilities(ctx context.Context) error
	GetUserQuota(ctx context.Context, userID int64) (int64, error)
	EnsureGroup(ctx context.Context, group, displayName string) error
}
