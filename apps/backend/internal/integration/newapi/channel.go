package newapi

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

const channelAddModeSingle = "single"

type addChannelRequest struct {
	Mode    string            `json:"mode"`
	Channel UpsertChannelBody `json:"channel"`
}

// UpsertChannelBody is the JSON body for PUT /api/channel/ (full replacement).
type UpsertChannelBody struct {
	ID     int    `json:"id,omitempty"`
	Type   int    `json:"type"`
	Name   string `json:"name"`
	Key    string `json:"key"`
	Status int    `json:"status"`
	Group  string `json:"group,omitempty"`
}

func (c *Client) UpsertChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	if req.ID > 0 {
		return c.updateChannel(ctx, req)
	}
	return c.createChannel(ctx, req)
}

func (c *Client) createChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	payload := addChannelRequest{
		Mode: channelAddModeSingle,
		Channel: UpsertChannelBody{
			Type:   req.Type,
			Name:   req.Name,
			Key:    req.Key,
			Status: req.Status,
			Group:  req.Group,
		},
	}
	// Upstream AddChannel returns success with empty data; resolve by unique name.
	if err := c.do(ctx, "POST", "/api/channel/", payload, nil); err != nil {
		return adminport.ChannelResult{}, err
	}
	ch, err := c.findChannelByName(ctx, req.Name)
	if err != nil {
		return adminport.ChannelResult{}, err
	}
	if ch.ID <= 0 {
		return adminport.ChannelResult{}, fmt.Errorf("create channel succeeded but id missing for %q", req.Name)
	}
	return adminport.ChannelResult{ID: ch.ID}, nil
}

func (c *Client) updateChannel(ctx context.Context, req adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
	cur, err := c.GetChannel(ctx, req.ID)
	if err != nil {
		return adminport.ChannelResult{}, err
	}
	body := MergeChannelPut(cur, req)
	var out Channel
	if err := c.do(ctx, "PUT", "/api/channel/", body, &out); err != nil {
		return adminport.ChannelResult{}, err
	}
	id := out.ID
	if id == 0 {
		id = req.ID
	}
	return adminport.ChannelResult{ID: id}, nil
}

func MergeChannelPut(cur Channel, req adminport.UpsertChannelInput) UpsertChannelBody {
	return UpsertChannelBody{
		ID:     req.ID,
		Type:   coalesceNonZeroInt(req.Type, cur.Type),
		Name:   coalesceString(req.Name, cur.Name),
		Key:    coalesceString(req.Key, cur.Key),
		Status: coalesceNonZeroInt(req.Status, cur.Status),
		Group:  coalesceString(req.Group, cur.Group),
	}
}

func (c *Client) GetChannel(ctx context.Context, channelID int) (Channel, error) {
	var ch Channel
	path := "/api/channel/" + strconv.Itoa(channelID)
	if err := c.do(ctx, "GET", path, nil, &ch); err != nil {
		return Channel{}, err
	}
	return ch, nil
}

func (c *Client) findChannelByName(ctx context.Context, name string) (Channel, error) {
	ch, err := findLatestByName(
		ctx, c, name, channelListFirstPage,
		func(page int) string {
			return "/api/channel/?p=" + strconv.Itoa(page) + "&page_size=100"
		},
		func(ch Channel) string { return ch.Name },
		func(ch Channel) int64 { return int64(ch.ID) },
	)
	if err != nil {
		return Channel{}, fmt.Errorf("newapi channel not found after create: %w", err)
	}
	return ch, nil
}

func (c *Client) RebuildAbilities(ctx context.Context) error {
	return c.do(ctx, "POST", "/api/channel/fix", nil, nil)
}
