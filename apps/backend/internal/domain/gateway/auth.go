package gateway

import "strings"

func parseBearerSecret(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	secret := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if secret == "" {
		return "", false
	}
	return secret, true
}
