package cryptoutil_test

import (
	"bytes"
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/cryptoutil"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := cryptoutil.DevDefaultKey()
	plaintext := []byte(`{"appId":"cli_test","appSecret":"secret"}`)
	encrypted, err := cryptoutil.Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := cryptoutil.Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestParseKeyAcceptsBase64(t *testing.T) {
	key, err := cryptoutil.ParseKey("dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
}
