package baseurl_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/baseurl"
)

func TestOrigin(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"http://localhost:3000", "http://localhost:3000"},
		{"http://localhost:3000/", "http://localhost:3000"},
		{"http://localhost:3000/v1", "http://localhost:3000"},
		{"http://localhost:3000/v1/", "http://localhost:3000"},
		{" https://api.example.com/v1 ", "https://api.example.com"},
	}
	for _, tc := range cases {
		got, err := baseurl.Origin(tc.in)
		if err != nil {
			t.Fatalf("Origin(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("Origin(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestOriginRejects(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "ftp://x", "http:///nohost", "http://localhost:3000/api", "http://localhost:3000/v1/extra"} {
		if _, err := baseurl.Origin(in); err == nil {
			t.Fatalf("Origin(%q) expected error", in)
		}
	}
}
