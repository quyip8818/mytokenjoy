package redemption_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/redemption"
)

func TestGenerateCodeFormat(t *testing.T) {
	t.Parallel()
	code, err := redemption.GenerateCode()
	if err != nil {
		t.Fatal(err)
	}
	// Format: TJ-XXXX-XXXX-XXXX (17 chars total: 2 prefix + 3 dashes + 12 body)
	if len(code) != 17 {
		t.Fatalf("code length: got %d want 17 (%q)", len(code), code)
	}
	if !strings.HasPrefix(code, "TJ-") {
		t.Fatalf("code must start with TJ-: %q", code)
	}
	parts := strings.Split(code, "-")
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts separated by -, got %d: %q", len(parts), code)
	}
	if parts[0] != "TJ" {
		t.Fatalf("prefix: got %q want TJ", parts[0])
	}
	for _, p := range parts[1:] {
		if len(p) != 4 {
			t.Fatalf("each segment should be 4 chars: %q in %q", p, code)
		}
	}
}

func TestGenerateCodeUniqueness(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		code, err := redemption.GenerateCode()
		if err != nil {
			t.Fatal(err)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("duplicate code after %d generations: %q", i, code)
		}
		seen[code] = struct{}{}
	}
}

func TestGenerateCodeCharset(t *testing.T) {
	t.Parallel()
	const allowed = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	for i := 0; i < 50; i++ {
		code, err := redemption.GenerateCode()
		if err != nil {
			t.Fatal(err)
		}
		body := strings.ReplaceAll(code[3:], "-", "") // strip "TJ-" prefix and dashes
		for _, ch := range body {
			if !strings.ContainsRune(allowed, ch) {
				t.Fatalf("invalid char %q in code %q", string(ch), code)
			}
		}
	}
}

func TestNormalizeCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"TJ-A3B4-C5D6-E7F8", "TJ-A3B4-C5D6-E7F8"},
		{"tj-a3b4-c5d6-e7f8", "TJ-A3B4-C5D6-E7F8"},
		{"TJA3B4C5D6E7F8", "TJ-A3B4-C5D6-E7F8"},
		{"  tj a3b4 c5d6 e7f8  ", "TJ-A3B4-C5D6-E7F8"},
		{"A3B4C5D6E7F8", "TJ-A3B4-C5D6-E7F8"}, // no prefix
		{"too-short", "too-short"},               // can't normalize, pass through
	}
	for _, tt := range tests {
		got := redemption.NormalizeCode(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
