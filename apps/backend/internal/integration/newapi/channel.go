package newapi

import (
	"context"
	"strings"
)

const (
	ChannelStatusEnabled  = 1
	ChannelStatusDisabled = 2
)

func ProviderChannelType(provider string) int {
	switch strings.ToLower(provider) {
	case "openai":
		return 1
	case "anthropic":
		return 14
	case "deepseek":
		return 25
	case "qwen":
		return 17
	case "azure":
		return 3
	default:
		return 1
	}
}

func (c *Client) UpsertChannel(ctx context.Context, req UpsertChannelRequest) (Channel, error) {
	method := "POST"
	path := "/api/channel/"
	if req.ID > 0 {
		method = "PUT"
	}
	var channel Channel
	if err := c.do(ctx, method, path, req, &channel); err != nil {
		return Channel{}, err
	}
	return channel, nil
}

func (c *Client) RebuildAbilities(ctx context.Context) error {
	return c.do(ctx, "GET", "/api/ability/sync", nil, nil)
}
