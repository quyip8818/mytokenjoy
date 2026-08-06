package invitetoken_test

import (
	"encoding/hex"
	"testing"

	"github.com/tokenjoy/backend/internal/support/invitetoken"
)

// Valid 32-byte hex key for tests.
const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()
	iss, err := invitetoken.NewIssuer(testKey)
	if err != nil {
		t.Fatal(err)
	}

	code := "deadbeef12345678" // 16 hex chars = 8 bytes
	for _, ch := range []string{invitetoken.ChannelSMS, invitetoken.ChannelEmail, invitetoken.ChannelAdminLink} {
		token, err := iss.Encrypt(code, ch)
		if err != nil {
			t.Fatalf("Encrypt(%s, %s): %v", code, ch, err)
		}
		if len(token) != 18 {
			t.Fatalf("expected 18-char token, got %d: %q", len(token), token)
		}

		payload, err := iss.Decrypt(token)
		if err != nil {
			t.Fatalf("Decrypt(%q): %v", token, err)
		}
		if payload.Code != code {
			t.Fatalf("code mismatch: want %s, got %s", code, payload.Code)
		}
		if payload.Channel != ch {
			t.Fatalf("channel mismatch: want %s, got %s", ch, payload.Channel)
		}
	}
}

func TestDecryptRejectsCorruptedToken(t *testing.T) {
	t.Parallel()
	iss, err := invitetoken.NewIssuer(testKey)
	if err != nil {
		t.Fatal(err)
	}

	token, err := iss.Encrypt("aaaaaaaabbbbbbbb", invitetoken.ChannelEmail)
	if err != nil {
		t.Fatal(err)
	}

	// Flip a byte in the token.
	raw := []byte(token)
	raw[0] ^= 0xff
	corrupted := string(raw)

	_, err = iss.Decrypt(corrupted)
	if err == nil {
		t.Fatal("expected error for corrupted token")
	}
}

func TestKeyRotationDecryptsWithOldKey(t *testing.T) {
	t.Parallel()
	oldKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newKey := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	oldIss, err := invitetoken.NewIssuer(oldKey)
	if err != nil {
		t.Fatal(err)
	}

	// Encrypt with old key.
	token, err := oldIss.Encrypt("1122334455667788", invitetoken.ChannelSMS)
	if err != nil {
		t.Fatal(err)
	}

	// New issuer with newKey primary, oldKey as fallback.
	rotatedIss, err := invitetoken.NewIssuer(newKey, oldKey)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := rotatedIss.Decrypt(token)
	if err != nil {
		t.Fatalf("rotated issuer should decrypt old-key token: %v", err)
	}
	if payload.Code != "1122334455667788" {
		t.Fatalf("code mismatch: %s", payload.Code)
	}
}

func TestNewIssuerRejectsInvalidKey(t *testing.T) {
	t.Parallel()
	// Too short.
	_, err := invitetoken.NewIssuer("aabb")
	if err == nil {
		t.Fatal("expected error for short key")
	}

	// Not hex.
	_, err = invitetoken.NewIssuer("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	if err == nil {
		t.Fatal("expected error for non-hex key")
	}

	// No keys at all.
	_, err = invitetoken.NewIssuer()
	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}

func TestEncryptRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	iss, err := invitetoken.NewIssuer(testKey)
	if err != nil {
		t.Fatal(err)
	}

	// Code too short.
	_, err = iss.Encrypt("aabb", "s")
	if err == nil {
		t.Fatal("expected error for short code")
	}

	// Channel too long.
	_, err = iss.Encrypt("deadbeef12345678", "sms")
	if err == nil {
		t.Fatal("expected error for multi-char channel")
	}

	// Invalid hex code.
	_, err = iss.Encrypt("zzzzzzzzzzzzzzzz", "e")
	if err == nil {
		t.Fatal("expected error for non-hex code")
	}
}

func TestTokenDeterministic(t *testing.T) {
	t.Parallel()
	iss, err := invitetoken.NewIssuer(testKey)
	if err != nil {
		t.Fatal(err)
	}

	// Same input → same token (XOR scheme is deterministic, no random nonce).
	code := "0000000000000000"
	t1, _ := iss.Encrypt(code, invitetoken.ChannelAdminLink)
	t2, _ := iss.Encrypt(code, invitetoken.ChannelAdminLink)
	if t1 != t2 {
		t.Fatalf("expected deterministic tokens, got %q vs %q", t1, t2)
	}

	// Different channel → different token.
	t3, _ := iss.Encrypt(code, invitetoken.ChannelSMS)
	if t1 == t3 {
		t.Fatal("different channels should produce different tokens")
	}
}

// Sanity check: 16 hex chars = 8 bytes.
func TestCodeHexLength(t *testing.T) {
	t.Parallel()
	b, _ := hex.DecodeString("deadbeef12345678")
	if len(b) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(b))
	}
}
