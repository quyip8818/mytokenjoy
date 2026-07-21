package newapi

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

// SelfHealingPort wraps a Client and automatically refreshes the admin token
// when a 401 Unauthorized response is detected.
type SelfHealingPort struct {
	client     *Client
	tokenStore *TokenStore
	mu         sync.Mutex
}

var _ adminport.Port = (*SelfHealingPort)(nil)

// NewSelfHealingPort creates a Port that auto-refreshes its token on 401.
func NewSelfHealingPort(client *Client, tokenStore *TokenStore) *SelfHealingPort {
	return &SelfHealingPort{client: client, tokenStore: tokenStore}
}

func (p *SelfHealingPort) selfHeal(ctx context.Context, err error) bool {
	if !isUnauthorized(err) {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	newToken, fetchErr := p.tokenStore.FetchToken(ctx)
	if fetchErr != nil {
		slog.Warn("newapi self-heal: failed to refresh token", "error", fetchErr)
		return false
	}
	if newToken == p.client.adminToken {
		// Token unchanged in DB — nothing to heal.
		return false
	}
	p.client.adminToken = newToken
	slog.Info("newapi self-heal: token refreshed")
	return true
}

func isUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "Unauthorized") ||
		strings.Contains(msg, "not logged in")
}

// --- adminport.Port delegation with self-healing ---

func (p *SelfHealingPort) CreateToken(ctx context.Context, req adminport.CreateTokenInput) (adminport.TokenResult, error) {
	r, err := p.client.CreateToken(ctx, req)
	if p.selfHeal(ctx, err) {
		return p.client.CreateToken(ctx, req)
	}
	return r, err
}

func (p *SelfHealingPort) UpdateToken(ctx context.Context, req adminport.UpdateTokenInput) (adminport.TokenResult, error) {
	r, err := p.client.UpdateToken(ctx, req)
	if p.selfHeal(ctx, err) {
		return p.client.UpdateToken(ctx, req)
	}
	return r, err
}

func (p *SelfHealingPort) GetToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	r, err := p.client.GetToken(ctx, tokenID)
	if p.selfHeal(ctx, err) {
		return p.client.GetToken(ctx, tokenID)
	}
	return r, err
}

func (p *SelfHealingPort) GetTokenKey(ctx context.Context, tokenID int64) (string, error) {
	r, err := p.client.GetTokenKey(ctx, tokenID)
	if p.selfHeal(ctx, err) {
		return p.client.GetTokenKey(ctx, tokenID)
	}
	return r, err
}

func (p *SelfHealingPort) RegenerateToken(ctx context.Context, tokenID int64) (adminport.TokenResult, error) {
	r, err := p.client.RegenerateToken(ctx, tokenID)
	if p.selfHeal(ctx, err) {
		return p.client.RegenerateToken(ctx, tokenID)
	}
	return r, err
}

func (p *SelfHealingPort) DeleteToken(ctx context.Context, tokenID int64) error {
	err := p.client.DeleteToken(ctx, tokenID)
	if p.selfHeal(ctx, err) {
		return p.client.DeleteToken(ctx, tokenID)
	}
	return err
}

func (p *SelfHealingPort) UpsertChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	r, err := p.client.UpsertChannel(ctx, req)
	if p.selfHeal(ctx, err) {
		return p.client.UpsertChannel(ctx, req)
	}
	return r, err
}

func (p *SelfHealingPort) EnsureGroup(ctx context.Context, group, displayName string) error {
	err := p.client.EnsureGroup(ctx, group, displayName)
	if p.selfHeal(ctx, err) {
		return p.client.EnsureGroup(ctx, group, displayName)
	}
	return err
}

func (p *SelfHealingPort) RebuildAbilities(ctx context.Context) error {
	err := p.client.RebuildAbilities(ctx)
	if p.selfHeal(ctx, err) {
		return p.client.RebuildAbilities(ctx)
	}
	return err
}

func (p *SelfHealingPort) CreateUser(ctx context.Context, req adminport.CreateUserInput) (adminport.UserResult, error) {
	r, err := p.client.CreateUser(ctx, req)
	if p.selfHeal(ctx, err) {
		return p.client.CreateUser(ctx, req)
	}
	return r, err
}

func (p *SelfHealingPort) ListModelPricing(ctx context.Context) ([]adminport.ModelPricing, error) {
	r, err := p.client.ListModelPricing(ctx)
	if p.selfHeal(ctx, err) {
		return p.client.ListModelPricing(ctx)
	}
	return r, err
}

func (p *SelfHealingPort) UpdateOption(ctx context.Context, key, value string) error {
	err := p.client.UpdateOption(ctx, key, value)
	if p.selfHeal(ctx, err) {
		return p.client.UpdateOption(ctx, key, value)
	}
	return err
}

func (p *SelfHealingPort) UpsertModelRatio(ctx context.Context, modelType string, inputPrice, outputPrice float64) error {
	err := p.client.UpsertModelRatio(ctx, modelType, inputPrice, outputPrice)
	if p.selfHeal(ctx, err) {
		return p.client.UpsertModelRatio(ctx, modelType, inputPrice, outputPrice)
	}
	return err
}

// FormatError wraps a NewAPI token initialization error with actionable context.
func FormatError(err error) error {
	return fmt.Errorf("newapi admin token: %w — ensure NewAPI is running and its database is accessible", err)
}
