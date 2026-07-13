package newapi

import (
	"context"
)

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
	return c.do(ctx, "GET", "/api/channel/sync", nil, nil)
}
