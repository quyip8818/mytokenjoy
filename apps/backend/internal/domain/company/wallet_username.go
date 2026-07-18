package company

import (
	"math/big"

	"github.com/google/uuid"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// WalletUsername derives a deterministic NewAPI wallet username from a company UUID.
// It encodes the UUID as base62 and returns the last 20 characters (≤119-bit space).
// NewAPI (QuantumNous/new-api) enforces validate:"max=20" on Username.
func WalletUsername(id uuid.UUID) string {
	n := new(big.Int).SetBytes(id[:])
	encoded := encodeBase62(n)
	if len(encoded) <= 20 {
		return encoded
	}
	return encoded[len(encoded)-20:]
}

func encodeBase62(n *big.Int) string {
	if n.Sign() == 0 {
		return "0"
	}
	base := big.NewInt(62)
	mod := new(big.Int)
	buf := make([]byte, 0, 22) // UUID base62 is at most 22 chars
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		buf = append(buf, base62Chars[mod.Int64()])
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
