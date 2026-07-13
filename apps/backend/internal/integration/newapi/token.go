package newapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (c *Client) CreateToken(ctx context.Context, req CreateTokenRequest) (Token, error) {
	if err := c.do(ctx, "POST", "/api/token/", req, nil); err != nil {
		return Token{}, err
	}
	token, err := c.findTokenByName(ctx, req.Name)
	if err != nil {
		return Token{}, err
	}
	if token.Key == "" || strings.Contains(token.Key, "*") {
		return c.RegenerateToken(ctx, token.ID)
	}
	return token, nil
}

type tokenListPage struct {
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Total    int     `json:"total"`
	Items    []Token `json:"items"`
}

func (c *Client) findTokenByName(ctx context.Context, name string) (Token, error) {
	var best Token
	var found bool
	for page := 0; page < 20; page++ {
		var list tokenListPage
		path := "/api/token/?p=" + strconv.Itoa(page)
		if err := c.do(ctx, "GET", path, nil, &list); err != nil {
			return Token{}, err
		}
		for _, item := range list.Items {
			if item.Name == name && (!found || item.ID > best.ID) {
				best = item
				found = true
			}
		}
		if len(list.Items) == 0 {
			break
		}
		pageSize := list.PageSize
		if pageSize <= 0 {
			pageSize = len(list.Items)
		}
		if list.Total > 0 && (page+1)*pageSize >= list.Total {
			break
		}
	}
	if !found {
		return Token{}, fmt.Errorf("newapi token not found after create: %s", name)
	}
	return best, nil
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

type tokenKeyResponse struct {
	Key string `json:"key"`
}

// GetTokenKey returns the full platform key secret without rotating it.
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
