package baseurl

import (
	"fmt"
	"net/url"
	"strings"
)

// Origin returns scheme://host[:port] for clients that append "/v1/..." themselves
// (NewAPI admin/gateway and OpenAI-compatible channels). Strips trailing slash
// and a lone "/v1" API root; rejects any other path.
func Origin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("url scheme must be http or https")
	}
	if u.Host == "" {
		return "", fmt.Errorf("url host is required")
	}
	path := strings.TrimSuffix(u.EscapedPath(), "/")
	if path == "/v1" {
		path = ""
	}
	if path != "" {
		return "", fmt.Errorf("base url must not include a path (got %q)", u.Path)
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
