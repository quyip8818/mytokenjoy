package newapi

func coalesceString(override, current string) string {
	if override != "" {
		return override
	}
	return current
}

func coalescePtr[T any](override *T, current T) T {
	if override != nil {
		return *override
	}
	return current
}

func coalesceNonZeroInt(override, current int) int {
	if override != 0 {
		return override
	}
	return current
}
