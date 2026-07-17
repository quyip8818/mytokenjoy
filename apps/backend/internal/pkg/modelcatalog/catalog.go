package modelcatalog

import (
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

var (
	ErrUnknownModelID = errors.New("unknown model id")
	ErrModelDisabled  = errors.New("model disabled")
)

func DedupeEffective(items []types.ModelInfo) []types.ModelInfo {
	byKey := make(map[string]types.ModelInfo, len(items))
	order := make([]string, 0, len(items))
	for _, item := range items {
		key := types.ModelCatalogKey(item.Provider, item.Type)
		if _, seen := byKey[key]; !seen {
			order = append(order, key)
		}
		byKey[key] = item
	}
	out := make([]types.ModelInfo, 0, len(order))
	for _, key := range order {
		out = append(out, byKey[key])
	}
	return out
}

func IndexByID(catalog []types.ModelInfo) map[uuid.UUID]types.ModelInfo {
	byID := make(map[uuid.UUID]types.ModelInfo, len(catalog))
	for _, item := range catalog {
		byID[item.ID] = item
	}
	return byID
}

func IndexByCallType(catalog []types.ModelInfo) map[string][]types.ModelInfo {
	byType := make(map[string][]types.ModelInfo)
	for _, item := range catalog {
		byType[item.Type] = append(byType[item.Type], item)
	}
	return byType
}

func FilterEnabledIDs(catalog []types.ModelInfo, ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	byID := IndexByID(catalog)
	out := make([]uuid.UUID, 0, len(ids))
	for _, mid := range ids {
		item, ok := byID[mid]
		if !ok || !item.Enabled {
			continue
		}
		out = append(out, mid)
	}
	return out
}

func FilterValidIDs(catalog []types.ModelInfo, ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	byID := IndexByID(catalog)
	out := make([]uuid.UUID, 0, len(ids))
	for _, mid := range ids {
		if _, ok := byID[mid]; ok {
			out = append(out, mid)
		}
	}
	return out
}

func IsCallTypeAllowed(catalog []types.ModelInfo, allowedIDs []uuid.UUID, callType string) bool {
	if callType == "" {
		return false
	}
	byID := IndexByID(catalog)
	for _, mid := range allowedIDs {
		item, ok := byID[mid]
		if !ok || !item.Enabled {
			continue
		}
		if item.Type == callType {
			return true
		}
	}
	return false
}

func ResolveIDForCallType(catalog []types.ModelInfo, allowedIDs []uuid.UUID, callType string) (*uuid.UUID, bool) {
	if callType == "" || len(allowedIDs) == 0 {
		return nil, false
	}
	byID := IndexByID(catalog)
	matches := make([]types.ModelInfo, 0)
	for _, mid := range allowedIDs {
		item, ok := byID[mid]
		if !ok || !item.Enabled || item.Type != callType {
			continue
		}
		matches = append(matches, item)
	}
	if len(matches) == 0 {
		return nil, false
	}
	if len(matches) == 1 {
		id := matches[0].ID
		return &id, true
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Provider == types.ProviderCustom && matches[j].Provider != types.ProviderCustom {
			return true
		}
		if matches[i].Provider != types.ProviderCustom && matches[j].Provider == types.ProviderCustom {
			return false
		}
		return matches[i].Provider < matches[j].Provider
	})
	id := matches[0].ID
	return &id, true
}

func CallTypesForIDs(catalog []types.ModelInfo, ids []uuid.UUID) []string {
	if len(ids) == 0 {
		return nil
	}
	byID := IndexByID(catalog)
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, mid := range ids {
		item, ok := byID[mid]
		if !ok || !item.Enabled {
			continue
		}
		if _, ok := seen[item.Type]; ok {
			continue
		}
		seen[item.Type] = struct{}{}
		out = append(out, item.Type)
	}
	sort.Strings(out)
	return out
}

func ToModelRef(item types.ModelInfo) types.ModelRef {
	return types.ModelRef{
		ID:       item.ID,
		Type:     item.Type,
		Name:     item.Name,
		Provider: item.Provider,
		Enabled:  item.Enabled,
	}
}

func EnrichRefs(catalog []types.ModelInfo, ids []uuid.UUID) []types.ModelRef {
	if len(ids) == 0 {
		return nil
	}
	byID := IndexByID(catalog)
	out := make([]types.ModelRef, 0, len(ids))
	for _, mid := range ids {
		item, ok := byID[mid]
		if !ok {
			continue
		}
		out = append(out, ToModelRef(item))
	}
	return out
}

func EnrichRef(catalog []types.ModelInfo, modelPtr *uuid.UUID) *types.ModelRef {
	if modelPtr == nil {
		return nil
	}
	byID := IndexByID(catalog)
	item, ok := byID[*modelPtr]
	if !ok {
		return nil
	}
	ref := ToModelRef(item)
	return &ref
}

func ValidateWritableIDs(catalog []types.ModelInfo, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	byID := IndexByID(catalog)
	for _, mid := range ids {
		item, ok := byID[mid]
		if !ok {
			return ErrUnknownModelID
		}
		if !item.Enabled {
			return ErrModelDisabled
		}
	}
	return nil
}

func EnabledModelIDs(catalog []types.ModelInfo) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(catalog))
	for _, item := range catalog {
		if item.Enabled {
			out = append(out, item.ID)
		}
	}
	return out
}
