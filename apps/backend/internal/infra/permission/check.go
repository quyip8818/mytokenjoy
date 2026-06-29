package permission

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
