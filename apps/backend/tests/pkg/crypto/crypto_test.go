package crypto_test

import (
	"bytes"
	"testing"

	"github.com/tokenjoy/backend/internal/support/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()
	key := crypto.DevDefaultKey()
	plaintext := []byte(`{"appId":"cli_test","appSecret":"secret"}`)
	encrypted, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := crypto.Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestParseKeyAcceptsBase64(t *testing.T) {
	t.Parallel()
	key, err := crypto.ParseKey("dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
}

func TestEncryptDecryptFieldRoundTrip(t *testing.T) {
	t.Parallel()
	key := crypto.DevDefaultKey()
	encrypted, err := crypto.EncryptField(key, "sk-provider-secret")
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.IsEncryptedField(encrypted) {
		t.Fatalf("expected encrypted prefix, got %q", encrypted)
	}
	plain, err := crypto.DecryptField(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "sk-provider-secret" {
		t.Fatalf("expected round-trip plaintext, got %q", plain)
	}
}
