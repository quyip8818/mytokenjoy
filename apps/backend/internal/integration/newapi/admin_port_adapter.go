package newapi

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

type AdminPortAdapter struct {
	client AdminClient
}

func NewAdminPortAdapter(client AdminClient) adminport.Port {
	if client == nil {
		return nil
	}
	return AdminPortAdapter{client: client}
}

func mapTokenResult(token Token, err error) (adminport.TokenResult, error) {
	if err != nil {
		return adminport.TokenResult{}, err
	}
	return adminport.TokenResult{ID: token.ID, Key: token.Key, RemainQuota: token.RemainQuota}, nil
}

func (a AdminPortAdapter) CreateToken(ctx context.Context, req adminport.CreateTokenInput) (adminport.TokenResult, error) {
	return mapTokenResult(a.client.CreateToken(ctx, CreateTokenRequest{
		UserID:             req.UserID,
		Name:               req.Name,
		RemainQuota:        req.RemainQuota,
		UnlimitedQuota:     req.UnlimitedQuota,
		ModelLimitsEnabled: req.ModelLimitsEnabled,
		ModelLimits:        req.ModelLimits,
		Group:              req.Group,
		ExpiredTime:        req.ExpiredTime,
	}))
}

func (a AdminPortAdapter) UpdateToken(ctx context.Context, req adminport.UpdateTokenInput) (adminport.TokenResult, error) {
	return mapTokenResult(a.client.UpdateToken(ctx, UpdateTokenRequest{
		ID:                 req.ID,
		Name:               req.Name,
		Status:             req.Status,
		RemainQuota:        req.RemainQuota,
		UnlimitedQuota:     req.UnlimitedQuota,
		ModelLimitsEnabled: req.ModelLimitsEnabled,
		ModelLimits:        req.ModelLimits,
		Group:              req.Group,
	}))
}

func (a AdminPortAdapter) GetToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	return mapTokenResult(a.client.GetToken(ctx, tokenID))
}

func (a AdminPortAdapter) GetTokenKey(ctx context.Context, tokenID int64) (string, error) {
	return a.client.GetTokenKey(ctx, tokenID)
}

func (a AdminPortAdapter) RegenerateToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	return mapTokenResult(a.client.RegenerateToken(ctx, tokenID))
}

func (a AdminPortAdapter) DeleteToken(ctx context.Context, tokenID int64) error {
	return a.client.DeleteToken(ctx, tokenID)
}

func (a AdminPortAdapter) UpsertChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	channel, err := a.client.UpsertChannel(ctx, UpsertChannelRequest{
		ID:     req.ID,
		Type:   req.Type,
		Name:   req.Name,
		Key:    req.Key,
		Status: req.Status,
	})
	if err != nil {
		return adminport.ChannelResult{}, err
	}
	return adminport.ChannelResult{ID: channel.ID}, nil
}

func (a AdminPortAdapter) CreateUser(ctx context.Context, req adminport.CreateUserInput) (adminport.UserResult, error) {
	user, err := a.client.CreateUser(ctx, CreateUserRequest{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		Quota:       req.Quota,
	})
	if err != nil {
		return adminport.UserResult{}, err
	}
	return adminport.UserResult{ID: user.ID}, nil
}

func (a AdminPortAdapter) TopUp(ctx context.Context, req adminport.TopUpInput) error {
	return a.client.TopUp(ctx, TopUpRequest{
		UserID: req.UserID,
		Quota:  req.Quota,
		Remark: req.Remark,
	})
}

func (a AdminPortAdapter) RebuildAbilities(ctx context.Context) error {
	return a.client.RebuildAbilities(ctx)
}

func (a AdminPortAdapter) GetUserQuota(ctx context.Context, userID int64) (int64, error) {
	return a.client.GetUserQuota(ctx, userID)
}
