package avatar_test

import (
	"encoding/base64"
	"strings"
	"testing"

	mehandler "github.com/tokenjoy/backend/internal/http/handler/me"
)

func TestValidateAvatar_Empty(t *testing.T) {
	t.Parallel()
	if err := mehandler.ValidateAvatar(""); err != nil {
		t.Fatalf("empty avatar should be valid, got: %v", err)
	}
}

func TestValidateAvatar_DiceBearValid(t *testing.T) {
	t.Parallel()
	cases := []string{
		"dicebear:adventurer:seed123",
		"dicebear:notionists:abc",
		"dicebear:bottts:x",
		"dicebear:shapes:hello-world",
		"dicebear:lorelei:test",
		"dicebear:fun-emoji:foo",
	}
	for _, c := range cases {
		if err := mehandler.ValidateAvatar(c); err != nil {
			t.Errorf("expected valid for %q, got: %v", c, err)
		}
	}
}

func TestValidateAvatar_DiceBearInvalidStyle(t *testing.T) {
	t.Parallel()
	if err := mehandler.ValidateAvatar("dicebear:unknown:seed"); err == nil {
		t.Fatal("expected error for unsupported style")
	}
}

func TestValidateAvatar_DiceBearSeedTooLong(t *testing.T) {
	t.Parallel()
	longSeed := strings.Repeat("a", 65)
	if err := mehandler.ValidateAvatar("dicebear:adventurer:" + longSeed); err == nil {
		t.Fatal("expected error for seed > 64 chars")
	}
}

func TestValidateAvatar_DiceBearMalformed(t *testing.T) {
	t.Parallel()
	if err := mehandler.ValidateAvatar("dicebear:adventurer"); err == nil {
		t.Fatal("expected error for missing seed part")
	}
}

func TestValidateAvatar_DataURIValid(t *testing.T) {
	t.Parallel()
	data := base64.StdEncoding.EncodeToString([]byte{0xFF})
	avatar := "data:image/webp;base64," + data
	if err := mehandler.ValidateAvatar(avatar); err != nil {
		t.Fatalf("valid data URI should pass: %v", err)
	}
}

func TestValidateAvatar_DataURITooLarge(t *testing.T) {
	t.Parallel()
	bigData := make([]byte, 51*1024)
	encoded := base64.StdEncoding.EncodeToString(bigData)
	avatar := "data:image/webp;base64," + encoded
	if err := mehandler.ValidateAvatar(avatar); err == nil {
		t.Fatal("expected error for avatar > 50KB")
	}
}

func TestValidateAvatar_DataURIExactLimit(t *testing.T) {
	t.Parallel()
	data := make([]byte, 50*1024)
	encoded := base64.StdEncoding.EncodeToString(data)
	avatar := "data:image/webp;base64," + encoded
	if err := mehandler.ValidateAvatar(avatar); err != nil {
		t.Fatalf("50KB avatar should pass: %v", err)
	}
}

func TestValidateAvatar_DataURIInvalidBase64(t *testing.T) {
	t.Parallel()
	avatar := "data:image/webp;base64,!!!invalid!!!"
	if err := mehandler.ValidateAvatar(avatar); err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestValidateAvatar_DataURIMissingComma(t *testing.T) {
	t.Parallel()
	avatar := "data:image/webp;base64"
	if err := mehandler.ValidateAvatar(avatar); err == nil {
		t.Fatal("expected error for missing comma")
	}
}

func TestValidateAvatar_UnsupportedFormat(t *testing.T) {
	t.Parallel()
	if err := mehandler.ValidateAvatar("https://example.com/avatar.png"); err == nil {
		t.Fatal("expected error for URL format")
	}
	if err := mehandler.ValidateAvatar("random-string"); err == nil {
		t.Fatal("expected error for random string")
	}
}
