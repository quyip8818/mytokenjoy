package newapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// NewAPI uses -1 for never-expire; 0 means already expired.
const TokenExpiredNever int64 = -1

type TokenPutBody struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Status             int    `json:"status"`
	RemainQuota        int64  `json:"remain_quota"`
	UnlimitedQuota     bool   `json:"unlimited_quota"`
	ModelLimitsEnabled bool   `json:"model_limits_enabled"`
	ModelLimits        string `json:"model_limits"`
	Group              string `json:"group"`
	ExpiredTime        int64  `json:"expired_time"`
}

func (c *Client) CreateToken(ctx context.Context, req CreateTokenRequest) (Token, error) {
	var token Token
	if err := c.do(ctx, "POST", "/api/token/", req, &token); err != nil {
		return Token{}, err
	}
	if err := validateCreatedToken(req, token); err != nil {
		if token.ID > 0 {
			_ = c.DeleteToken(ctx, token.ID)
		}
		return Token{}, err
	}
	if token.Key == "" || strings.Contains(token.Key, "*") {
		return c.RegenerateToken(ctx, token.ID)
	}
	return token, nil
}

func validateCreatedToken(req CreateTokenRequest, token Token) error {
	if token.ID <= 0 {
		return fmt.Errorf("newapi create token: response missing id (admin create-token contract required)")
	}
	if req.UserID > 0 && token.UserID != req.UserID {
		return fmt.Errorf("newapi create token: owner mismatch want=%d got=%d", req.UserID, token.UserID)
	}
	return nil
}

func (c *Client) UpdateToken(ctx context.Context, req UpdateTokenRequest) (Token, error) {
	// NewAPI UpdateToken replaces the whole row; omitted JSON fields bind as zero
	// (expired_time=0 → immediately expired; empty name/group wipe platform metadata).
	cur, err := c.GetToken(ctx, req.ID)
	if err != nil {
		return Token{}, err
	}
	payload := MergeTokenPut(cur, req)
	var token Token
	if err := c.do(ctx, "PUT", "/api/token/", payload, &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func MergeTokenPut(cur Token, req UpdateTokenRequest) TokenPutBody {
	return TokenPutBody{
		ID:                 req.ID,
		Name:               coalesceString(req.Name, cur.Name),
		Status:             coalescePtr(req.Status, cur.Status),
		RemainQuota:        coalescePtr(req.RemainQuota, cur.RemainQuota),
		UnlimitedQuota:     coalescePtr(req.UnlimitedQuota, cur.UnlimitedQuota),
		ModelLimitsEnabled: coalescePtr(req.ModelLimitsEnabled, cur.ModelLimitsEnabled),
		ModelLimits:        coalesceString(req.ModelLimits, cur.ModelLimits),
		Group:              coalesceString(req.Group, cur.Group),
		ExpiredTime:        expiredTimeForPut(req.ExpiredTime, cur.ExpiredTime),
	}
}

// expiredTimeForPut preserves current expiry on partial updates. Upstream treats 0 as
// already-expired; heal legacy zero to never-expire when the caller did not override.
func expiredTimeForPut(override *int64, current int64) int64 {
	if override != nil {
		return *override
	}
	if current == 0 {
		return TokenExpiredNever
	}
	return current
}

func (c *Client) GetToken(ctx context.Context, tokenID int64) (Token, error) {
	var token Token
	path := "/api/token/" + strconv.FormatInt(tokenID, 10)
	if err := c.do(ctx, "GET", path, nil, &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

type tokenKeyResponse struct {
	Key string `json:"key"`
}

func (c *Client) GetTokenKey(ctx context.Context, tokenID int64) (string, error) {
	var out tokenKeyResponse
	path := "/api/token/" + strconv.FormatInt(tokenID, 10) + "/key"
	if err := c.do(ctx, "POST", path, nil, &out); err != nil {
		return "", err
	}
	return out.Key, nil
}

func (c *Client) DeleteToken(ctx context.Context, tokenID int64) error {
	path := "/api/token/" + strconv.FormatInt(tokenID, 10)
	return c.do(ctx, "DELETE", path, nil, nil)
}

func (c *Client) RegenerateToken(ctx context.Context, tokenID int64) (Token, error) {
	var token Token
	path := "/api/token/" + strconv.FormatInt(tokenID, 10) + "/regenerate"
	if err := c.do(ctx, "POST", path, nil, &token); err != nil {
		return Token{}, err
	}
	return token, nil
}
