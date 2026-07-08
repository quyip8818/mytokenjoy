package snapshot

import "github.com/tokenjoy/backend/seed/contract"

func modelIDByType() map[string]int64 {
	return map[string]int64{
		"gpt-4o":            contract.IDModel1,
		"gpt-4o-mini":       contract.IDModel2,
		"claude-opus-4-8":   contract.IDModel3,
		"claude-sonnet-4-6": contract.IDModel4,
		"deepseek-v3":       contract.IDModel5,
		"qwen-plus":         contract.IDModel8,
	}
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
