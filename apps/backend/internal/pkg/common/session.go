package common

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
)

const SessionCookie = "tokenjoy_session_member"

func UsedBearerAuth(r *http.Request) bool {
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		return false
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")) != ""
}

func ResolveMemberID(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token != "" {
			return token
		}
	}

	cookieHeader := r.Header.Get("Cookie")
	if cookieHeader != "" {
		re := regexp.MustCompile(SessionCookie + `=([^;]+)`)
		if match := re.FindStringSubmatch(cookieHeader); len(match) > 1 {
			return match[1]
		}
	}

	return ""
}

func ResolveDemoMemberName(memberID string, members []types.Member) string {
	if memberID == "" {
		return "审批人"
	}
	for _, member := range members {
		if member.ID == memberID {
			return member.Name
		}
	}
	return "审批人"
}
