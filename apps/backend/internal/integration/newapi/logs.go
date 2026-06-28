package newapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func (c *Client) ListLogs(ctx context.Context, params ListLogsParams) ([]LogEntry, error) {
	query := url.Values{}
	if params.Page > 0 {
		query.Set("p", strconv.Itoa(params.Page))
	}
	if params.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(params.PageSize))
	}
	if params.StartUnixTime > 0 {
		query.Set("start_timestamp", strconv.FormatInt(params.StartUnixTime, 10))
	}
	if params.EndUnixTime > 0 {
		query.Set("end_timestamp", strconv.FormatInt(params.EndUnixTime, 10))
	}
	path := "/api/log/"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var items []LogEntry
	if err := c.do(ctx, "GET", path, nil, &items); err != nil {
		return nil, err
	}
	if params.StartID > 0 {
		filtered := make([]LogEntry, 0, len(items))
		for _, item := range items {
			if item.ID > params.StartID {
				filtered = append(filtered, item)
			}
		}
		return filtered, nil
	}
	if items == nil {
		return []LogEntry{}, nil
	}
	return items, nil
}

func LogEntryModel(entry LogEntry) string {
	if entry.ModelName != "" {
		return entry.ModelName
	}
	return ""
}

func ValidateWebhookPayload(payload WebhookLogPayload) error {
	if payload.ID <= 0 {
		return fmt.Errorf("invalid log id")
	}
	if payload.TokenID <= 0 {
		return fmt.Errorf("invalid token id")
	}
	return nil
}
