package identity_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/verifycode"
)

func TestFormatPhone(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"13800138000", "+8613800138000"},
		{"+8613800138000", "+8613800138000"},
		{"+1234567890", "+1234567890"},
		{"86138", "86138"}, // too short, no prefix added
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := verifycode.FormatPhone(tc.input)
			if got != tc.want {
				t.Errorf("FormatPhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNewServiceNilWhenNoRedisURL(t *testing.T) {
	t.Parallel()
	svc, err := verifycode.NewService(verifycode.Config{RedisURL: ""}, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when redisURL is empty")
	}
}

func TestNewServiceErrorOnBadRedisURL(t *testing.T) {
	t.Parallel()
	svc, err := verifycode.NewService(verifycode.Config{RedisURL: "not-a-url"}, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid redis URL")
	}
	if svc != nil {
		t.Fatal("expected nil service on error")
	}
}
