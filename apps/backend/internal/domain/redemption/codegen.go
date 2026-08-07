package redemption

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// charset: 30 chars, excludes easily confused 0O1IL.
const charset = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// GenerateCode produces a single code in the format TJ-XXXX-XXXX-XXXX.
// Uses rejection sampling to avoid modulo bias (256 % 30 != 0).
func GenerateCode() (string, error) {
	// Read extra bytes to cover rejection (reject rate ~6.25%, so 14 bytes covers 12 with margin).
	buf := make([]byte, 16)
	code := make([]byte, 0, 12)
	for len(code) < 12 {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("crypto/rand: %w", err)
		}
		for _, b := range buf {
			if b < 240 {
				code = append(code, charset[b%30])
				if len(code) == 12 {
					break
				}
			}
		}
	}
	return fmt.Sprintf("TJ-%s-%s-%s", string(code[0:4]), string(code[4:8]), string(code[8:12])), nil
}

// NormalizeCode strips whitespace, removes dashes, uppercases, then re-formats.
// Accepts inputs like "tj-a3b4-c5d6-e7f8", "TJA3B4C5D6E7F8", "TJ A3B4 C5D6 E7F8".
func NormalizeCode(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	// Strip TJ prefix if present.
	s = strings.TrimPrefix(s, "TJ")
	if len(s) != 12 {
		return raw // can't normalize, return as-is for validation to reject
	}
	return fmt.Sprintf("TJ-%s-%s-%s", s[0:4], s[4:8], s[8:12])
}
