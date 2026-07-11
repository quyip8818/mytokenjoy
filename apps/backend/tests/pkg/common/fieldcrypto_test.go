package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestEncryptDecryptFieldRoundTrip(t *testing.T) {
	t.Parallel()
	key := common.DevDefaultKey()
	encrypted, err := common.EncryptField(key, "sk-provider-secret")
	if err != nil {
		t.Fatal(err)
	}
	if !common.IsEncryptedField(encrypted) {
		t.Fatalf("expected encrypted prefix, got %q", encrypted)
	}
	plain, err := common.DecryptField(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "sk-provider-secret" {
		t.Fatalf("expected round-trip plaintext, got %q", plain)
	}
}

func TestDecryptFieldLegacyPlaintext(t *testing.T) {
	t.Parallel()
	key := common.DevDefaultKey()
	plain, err := common.DecryptField(key, "legacy-plaintext-key")
	if err != nil {
		t.Fatal(err)
	}
	if plain != "legacy-plaintext-key" {
		t.Fatalf("expected legacy passthrough, got %q", plain)
	}
}
