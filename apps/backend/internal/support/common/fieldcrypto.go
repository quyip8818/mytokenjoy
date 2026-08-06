package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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
