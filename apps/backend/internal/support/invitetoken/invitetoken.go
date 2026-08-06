// Package invitetoken provides compact encrypted invite tokens (18 chars).
// Scheme: XOR encryption with HMAC-derived keystream + 4-byte truncated MAC.
// Plaintext: [8-byte code][1-byte channel] = 9 bytes.
// Token: base64url(9-byte ciphertext + 4-byte MAC) = 18 chars.
//
// ponytail: intentionally compact for SMS constraints (阿里云变量 ≤35 chars).
// Upgrade path: switch to AES-GCM if token length budget allows (50+ chars).
package invitetoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// Channel identifies how the invite was delivered.
const (
	ChannelSMS       = "s"
	ChannelEmail     = "e"
	ChannelAdminLink = "a"
)

// Payload is the decrypted invite token content.
type Payload struct {
	Code    string // hex-encoded invite code (for DB lookup)
	Channel string // single char: s/e/a
}

// Issuer encrypts and decrypts invite tokens.
type Issuer struct {
	keys [][]byte
}

// NewIssuer creates an Issuer from one or more hex-encoded 32-byte keys.
func NewIssuer(hexKeys ...string) (*Issuer, error) {
	if len(hexKeys) == 0 {
		return nil, errors.New("invitetoken: at least one key required")
	}
	keys := make([][]byte, 0, len(hexKeys))
	for _, h := range hexKeys {
		k, err := hex.DecodeString(h)
		if err != nil {
			return nil, fmt.Errorf("invitetoken: invalid hex key: %w", err)
		}
		if len(k) != 32 {
			return nil, fmt.Errorf("invitetoken: key must be 32 bytes, got %d", len(k))
		}
		keys = append(keys, k)
	}
	return &Issuer{keys: keys}, nil
}

// Encrypt produces an 18-char base64url token from invite code (16 hex chars) and channel.
func (iss *Issuer) Encrypt(codeHex, channel string) (string, error) {
	codeBytes, err := hex.DecodeString(codeHex)
	if err != nil || len(codeBytes) != 8 {
		return "", fmt.Errorf("invitetoken: code must be 16 hex chars, got %q", codeHex)
	}
	if len(channel) != 1 {
		return "", fmt.Errorf("invitetoken: channel must be single char, got %q", channel)
	}

	// plaintext = [8-byte code][1-byte channel]
	var plaintext [9]byte
	copy(plaintext[:8], codeBytes)
	plaintext[8] = channel[0]

	// Derive keystream: HMAC-SHA256(key, "enc") truncated to 9 bytes.
	keystream := deriveBlock(iss.keys[0], []byte("enc"))

	// XOR encrypt.
	var ciphertext [9]byte
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ keystream[i]
	}

	// MAC: HMAC-SHA256(key, ciphertext) truncated to 4 bytes.
	mac := computeMAC(iss.keys[0], ciphertext[:])

	// Output: ciphertext(9) + mac(4) = 13 bytes → base64url = 18 chars.
	var out [13]byte
	copy(out[:9], ciphertext[:])
	copy(out[9:], mac)
	return base64.RawURLEncoding.EncodeToString(out[:]), nil
}

// Decrypt decodes an 18-char token. Tries all keys for rotation support.
func (iss *Issuer) Decrypt(token string) (*Payload, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invitetoken: base64 decode: %w", err)
	}
	if len(raw) != 13 {
		return nil, fmt.Errorf("invitetoken: invalid token length %d", len(raw))
	}

	ciphertext := raw[:9]
	tokenMAC := raw[9:13]

	for _, key := range iss.keys {
		// Verify MAC first.
		expectedMAC := computeMAC(key, ciphertext)
		if !hmac.Equal(tokenMAC, expectedMAC) {
			continue
		}
		// Decrypt via XOR.
		keystream := deriveBlock(key, []byte("enc"))
		var plaintext [9]byte
		for i := range ciphertext {
			plaintext[i] = ciphertext[i] ^ keystream[i]
		}
		return &Payload{
			Code:    hex.EncodeToString(plaintext[:8]),
			Channel: string(plaintext[8:]),
		}, nil
	}
	return nil, errors.New("invitetoken: decrypt failed (bad MAC)")
}

// deriveBlock returns first 9 bytes of HMAC-SHA256(key, label).
func deriveBlock(key, label []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(label)
	return h.Sum(nil)[:9]
}

// computeMAC returns first 4 bytes of HMAC-SHA256(key, data).
func computeMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)[:4]
}
