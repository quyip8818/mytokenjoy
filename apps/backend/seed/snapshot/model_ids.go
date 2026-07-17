package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/seed/contract"
)

func modelIDByType() map[string]uuid.UUID {
	return contract.ModelTypeToID
}

func modelIDs(types []string) []uuid.UUID {
	index := modelIDByType()
	out := make([]uuid.UUID, 0, len(types))
	for _, modelType := range types {
		if id, ok := index[modelType]; ok {
			out = append(out, id)
		}
	}
	return out
}

func modelIDPtr(modelType string) *uuid.UUID {
	index := modelIDByType()
	if id, ok := index[modelType]; ok {
		return &id
	}
	return nil
}
