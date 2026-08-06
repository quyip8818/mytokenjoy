// Package crypto provides AES-GCM encryption primitives for credential fields.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const aesKeySize = 32

var ErrInvalidKey = errors.New("credential encryption key must be 32 bytes")

func ParseKey(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, ErrInvalidKey
	}
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(decoded) == aesKeySize {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(trimmed); err == nil && len(decoded) == aesKeySize {
		return decoded, nil
	}
	return nil, ErrInvalidKey
}

func DevDefaultKey() []byte {
	key, _ := ParseKey("dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=")
	return key
}

func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != aesKeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(key) != aesKeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, payload := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, payload, nil)
}

const encryptedFieldPrefix = "enc:v1:"

// EncryptField stores plaintext as an enc:v1:-prefixed base64 blob.
func EncryptField(key []byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if IsEncryptedField(plaintext) {
		return plaintext, nil
	}
	ciphertext, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return encryptedFieldPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptField returns plaintext for enc:v1: values.
func DecryptField(key []byte, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !IsEncryptedField(stored) {
		return "", fmt.Errorf("credential not encrypted")
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, encryptedFieldPrefix))
	if err != nil {
		return "", fmt.Errorf("decode encrypted field: %w", err)
	}
	plain, err := Decrypt(key, payload)
	if err != nil {
		return "", fmt.Errorf("decrypt field: %w", err)
	}
	return string(plain), nil
}

func IsEncryptedField(value string) bool {
	return strings.HasPrefix(value, encryptedFieldPrefix)
}
