package secrets

import (
	"crypto/rand"
	"encoding/hex"
)

// RandomHex returns n random bytes encoded as lowercase hex (2*n chars).
func RandomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
