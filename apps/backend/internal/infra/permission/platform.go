package permission

// IsPlatformPermission returns true for any permission in the platform domain (platform:*).
func IsPlatformPermission(p string) bool {
	const prefix = "platform:"
	return len(p) > len(prefix) && p[:len(prefix)] == prefix
}
