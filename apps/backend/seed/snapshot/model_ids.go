package snapshot

import "github.com/tokenjoy/backend/seed/contract"

func modelIDByType() map[string]int64 {
	return contract.ModelTypeToID
}

func modelIDs(types []string) []int64 {
	index := modelIDByType()
	out := make([]int64, 0, len(types))
	for _, modelType := range types {
		if id, ok := index[modelType]; ok {
			out = append(out, id)
		}
	}
	return out
}

func modelIDPtr(modelType string) *int64 {
	index := modelIDByType()
	if id, ok := index[modelType]; ok {
		return &id
	}
	return nil
}
