package newapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
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

// UpdateOption sets a NewAPI system option (e.g. ModelRatio, CompletionRatio).
func (c *Client) UpdateOption(ctx context.Context, key, value string) error {
	if err := c.do(ctx, "PUT", "/api/option/", map[string]string{
		"key":   key,
		"value": value,
	}, nil); err != nil {
		return fmt.Errorf("update option %s: %w", key, err)
	}
	return nil
}

// UpsertModelRatio adds or updates a single model's ratio entries in the
// global ModelRatio and CompletionRatio option maps (read-modify-write).
func (c *Client) UpsertModelRatio(ctx context.Context, modelType string, inputPrice, outputPrice float64) error {
	var entries []optionEntry
	if err := c.do(ctx, "GET", "/api/option/", nil, &entries); err != nil {
		return fmt.Errorf("list options for upsert ratio: %w", err)
	}
	byKey := make(map[string]string, len(entries))
	for _, e := range entries {
		byKey[e.Key] = e.Value
	}

	modelRatio, completionRatio := newapiunits.RatioFromPrice(inputPrice, outputPrice)

	// Update ModelRatio map
	mrMap := map[string]float64{}
	if raw := byKey["ModelRatio"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &mrMap)
	}
	mrMap[modelType] = modelRatio
	mrJSON, _ := json.Marshal(mrMap)
	if err := c.UpdateOption(ctx, "ModelRatio", string(mrJSON)); err != nil {
		return err
	}

	// Update CompletionRatio map
	crMap := map[string]float64{}
	if raw := byKey["CompletionRatio"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &crMap)
	}
	crMap[modelType] = completionRatio
	crJSON, _ := json.Marshal(crMap)
	if err := c.UpdateOption(ctx, "CompletionRatio", string(crJSON)); err != nil {
		return err
	}

	return nil
}
