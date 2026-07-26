// Package invitetoken provides AES-GCM encrypted invite tokens that embed
// the invite code, delivery channel, and expiry. The channel is tamper-proof
// because it lives inside authenticated ciphertext.
package invitetoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Channel identifies how the invite was delivered.
const (
	ChannelSMS       = "sms"
	ChannelEmail     = "email"
	ChannelAdminLink = "admin_link"
)

// Payload is the plaintext embedded in the encrypted token.
type Payload struct {
	Code    string `json:"code"`
	Channel string `json:"ch"`
	Exp     int64  `json:"exp"`
}

// Issuer encrypts and decrypts invite tokens.
type Issuer struct {
	keys [][]byte // first key is primary (encrypt), all are tried on decrypt
}

// NewIssuer creates an Issuer from one or more hex-encoded 32-byte keys.
// The first key is used for encryption; all keys are tried during decryption
// to support key rotation.
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

// Encrypt produces a base64url token embedding the given invite code, channel, and expiry.
func (iss *Issuer) Encrypt(code, channel string, expiresAt time.Time) (string, error) {
	payload := Payload{Code: code, Channel: channel, Exp: expiresAt.Unix()}
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("invitetoken: marshal: %w", err)
	}

	block, err := aes.NewCipher(iss.keys[0])
	if err != nil {
		return "", fmt.Errorf("invitetoken: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("invitetoken: gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("invitetoken: nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	raw := make([]byte, 0, len(nonce)+len(ciphertext))
	raw = append(raw, nonce...)
	raw = append(raw, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Decrypt decodes a base64url token and returns the payload.
// Tries all configured keys (supports rotation). Returns error if all fail or token is expired.
func (iss *Issuer) Decrypt(token string) (*Payload, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invitetoken: base64 decode: %w", err)
	}

	var lastErr error
	for _, key := range iss.keys {
		payload, err := decryptWithKey(key, raw)
		if err != nil {
			lastErr = err
			continue
		}
		// Check token-level expiry (DB expiry is checked separately).
		if time.Now().Unix() > payload.Exp {
			return nil, errors.New("invitetoken: token expired")
		}
		return payload, nil
	}
	return nil, fmt.Errorf("invitetoken: decrypt failed: %w", lastErr)
}

func decryptWithKey(key, raw []byte) (*Payload, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize+1 {
		return nil, errors.New("ciphertext too short")
	}
	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var payload Payload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("invitetoken: unmarshal: %w", err)
	}
	return &payload, nil
}
