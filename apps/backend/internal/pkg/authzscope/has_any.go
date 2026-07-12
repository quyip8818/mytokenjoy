package authzscope

// HasAny reports whether have includes any of required, or includes "*".
func HasAny(have []string, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	for _, p := range have {
		if p == "*" {
			return true
		}
	}
	for _, req := range required {
		if contains(have, req) {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
