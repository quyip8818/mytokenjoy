package usage

import "github.com/tokenjoy/backend/internal/store"

func ResolveConsumeModel(raw store.RawConsumeLog) string {
	return raw.ModelName
}
