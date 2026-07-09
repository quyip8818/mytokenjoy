package newapi

import (
	"context"
	"strconv"
)

func (c *Client) CreateToken(ctx context.Context, req CreateTokenRequest) (Token, error) {
	var token Token
	if err := c.do(ctx, "POST", "/api/token/", req, &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (c *Client) UpdateToken(ctx context.Context, req UpdateTokenRequest) (Token, error) {
	var token Token
	if err := c.do(ctx, "PUT", "/api/token/", req, &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (c *Client) GetToken(ctx context.Context, tokenID int64) (Token, error) {
	var token Token
	path := "/api/token/" + strconv.FormatInt(tokenID, 10)
	if err := c.do(ctx, "GET", path, nil, &token); err != nil {
		return Token{}, err
	}
	return token, nil
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
