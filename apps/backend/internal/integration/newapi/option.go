package newapi

import (
	"context"
	"encoding/json"
	"fmt"
)

type optionEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EnsureGroup registers group in NewAPI UserUsableGroups and GroupRatio if missing.
func (c *Client) EnsureGroup(ctx context.Context, group, displayName string) error {
	if group == "" {
		return nil
	}
	if displayName == "" {
		displayName = group
	}
	var entries []optionEntry
	if err := c.do(ctx, "GET", "/api/option/", nil, &entries); err != nil {
		return fmt.Errorf("list options: %w", err)
	}
	byKey := make(map[string]string, len(entries))
	for _, item := range entries {
		byKey[item.Key] = item.Value
	}
	for _, optKey := range []string{"UserUsableGroups", "GroupRatio"} {
		merged, skip, err := MergeGroupOption(byKey[optKey], optKey, group, displayName)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		if err := c.do(ctx, "PUT", "/api/option/", map[string]string{
			"key":   optKey,
			"value": merged,
		}, nil); err != nil {
			return fmt.Errorf("update option %s for group %s: %w", optKey, group, err)
		}
	}
	return nil
}

func MergeGroupOption(raw, optKey, group, displayName string) (merged string, skip bool, err error) {
	data := map[string]any{}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			return "", false, fmt.Errorf("parse option %s: %w", optKey, err)
		}
	}
	if _, ok := data[group]; ok {
		return "", true, nil
	}
	if optKey == "UserUsableGroups" {
		data[group] = displayName
	} else {
		data[group] = float64(1)
	}
	out, err := json.Marshal(data)
	if err != nil {
		return "", false, err
	}
	return string(out), false, nil
}
