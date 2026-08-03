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
	ListChannels(ctx context.Context) ([]ChannelInfo, error)
	DeleteChannel(ctx context.Context, channelID int) error
	EnsureGroup(ctx context.Context, group, displayName string) error
	RebuildAbilities(ctx context.Context) error

	// --- User provisioning (used by company creation, bootstrap) ---
	CreateUser(ctx context.Context, req CreateUserInput) (UserResult, error)
	ManageUser(ctx context.Context, userID int64, action string, value int64) error

	// --- Pricing (NewAPI cache sync) ---
	// ListModelPricing reads from NewAPI option store (debug/audit tooling).
	ListModelPricing(ctx context.Context) ([]ModelPricing, error)
	UpdateOption(ctx context.Context, key, value string) error
	// UpsertModelRatio pushes price→ratio to NewAPI (gateway pre-deduction cache).
	UpsertModelRatio(ctx context.Context, modelType string, inputPrice, outputPrice, cacheInputPrice float64) error
}
