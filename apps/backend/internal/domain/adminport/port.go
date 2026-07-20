package adminport

import "context"

type Port interface {
	// --- Token lifecycle (used by newapisync/platformkey) ---
	CreateToken(ctx context.Context, req CreateTokenInput) (TokenResult, error)
	UpdateToken(ctx context.Context, req UpdateTokenInput) (TokenResult, error)
	GetToken(ctx context.Context, tokenID int64) (TokenResult, error)
	GetTokenKey(ctx context.Context, tokenID int64) (string, error)
	RegenerateToken(ctx context.Context, tokenID int64) (TokenResult, error)
	DeleteToken(ctx context.Context, tokenID int64) error

	// --- Channel lifecycle (used by newapisync/provider) ---
	UpsertChannel(ctx context.Context, req UpsertChannelInput) (ChannelResult, error)
	EnsureGroup(ctx context.Context, group, displayName string) error
	RebuildAbilities(ctx context.Context) error

	// --- User provisioning (used by company creation, bootstrap) ---
	CreateUser(ctx context.Context, req CreateUserInput) (UserResult, error)

	// --- Pricing (used by models domain) ---
	ListModelPricing(ctx context.Context) ([]ModelPricing, error)
}
